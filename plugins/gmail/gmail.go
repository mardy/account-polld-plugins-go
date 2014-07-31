/*
 Copyright 2014 Canonical Ltd.
 Authors: Sergio Schvezov <sergio.schvezov@canonical.com>

 This program is free software: you can redistribute it and/or modify it
 under the terms of the GNU General Public License version 3, as published
 by the Free Software Foundation.

 This program is distributed in the hope that it will be useful, but
 WITHOUT ANY WARRANTY; without even the implied warranties of
 MERCHANTABILITY, SATISFACTORY QUALITY, or FITNESS FOR A PARTICULAR
 PURPOSE.  See the GNU General Public License for more details.

 You should have received a copy of the GNU General Public License along
 with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package gmail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
)

const (
	APP_ID = "com.ubuntu.developer.webapps.webapp-gmail_webapp-gmail"
)

var baseUrl, _ = url.Parse("https://www.googleapis.com/gmail/v1/users/me/")

type GmailPlugin struct {
	// reportedIds holds the messages that have already been notified. This
	// approach is taken against timestamps as it avoids needing to call
	// get on the message.
	//
	// TODO determine if persisting the list to avoid renotification on reboot.
	reportedIds map[string]bool
}

func New() *GmailPlugin {
	return &GmailPlugin{}
}

func (p *GmailPlugin) ApplicationId() plugins.ApplicationId {
	return plugins.ApplicationId(APP_ID)
}

func (p *GmailPlugin) Poll(authData *accounts.AuthData) ([]plugins.PushMessage, error) {
	// This envvar check is to ease testing.
	if token := os.Getenv("ACCOUNT_POLLD_TOKEN_GMAIL"); token != "" {
		authData.AccessToken = token
	}

	resp, err := p.requestMessageList(authData.AccessToken)
	if err != nil {
		return nil, err
	}
	messages, err := p.parseMessageListResponse(resp)
	if err != nil {
		return nil, err
	}

	// TODO use the batching API defined in https://developers.google.com/gmail/api/guides/batch
	for i := range messages {
		resp, err := p.requestMessage(messages[i].Id, authData.AccessToken)
		if err != nil {
			return nil, err
		}
		messages[i], err = p.parseMessageResponse(resp)
		if err != nil {
			return nil, err
		}
	}
	return p.createNotifications(messages)
}

func (p *GmailPlugin) reported(id string) bool {
	return p.reportedIds[id]
}

func (p *GmailPlugin) createNotifications(messages []message) ([]plugins.PushMessage, error) {
	pushMsgMap := make(pushes)

	for _, msg := range messages {
		hdr := msg.Payload.mapHeaders()
		if _, ok := pushMsgMap[msg.ThreadId]; ok {
			pushMsgMap[msg.ThreadId].Notification.Card.Summary += fmt.Sprintf(", %s", hdr[hdrFROM])
		} else {
			summary := fmt.Sprintf("Message \"%s\" from %s", hdr[hdrSUBJECT], hdr[hdrFROM])
			action := "https://mail.google.com/mail/u/0/?pli=1#inbox/" + msg.ThreadId
			pushMsgMap[msg.ThreadId] = *plugins.NewStandardPushMessage(summary, msg.Snippet, action, "")
		}
	}
	var pushMsg []plugins.PushMessage
	for _, v := range pushMsgMap {
		pushMsg = append(pushMsg, v)
	}
	return pushMsg, nil
}

func (p *GmailPlugin) parseMessageListResponse(resp *http.Response) ([]message, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp errorResp
		if err := decoder.Decode(&errResp); err != nil {
			return nil, err
		}
		if errResp.Err.Code == 401 {
			return nil, plugins.ErrTokenExpired
		}
		return nil, &errResp
	}

	var messages messageList
	if err := decoder.Decode(&messages); err != nil {
		return nil, err
	}

	filteredMsg := p.messageListFilter(messages.Messages)

	return filteredMsg, nil
}

// messageListFilter returns a subset of unread messages where the subset
// depends on not being in reportedIds. Before returning, reportedIds is
// updated with the new list of unread messages.
func (p *GmailPlugin) messageListFilter(messages []message) []message {
	sort.Sort(byId(messages))
	var reportMsg []message
	var ids = make(map[string]bool)

	for _, msg := range messages {
		if !p.reported(msg.Id) {
			reportMsg = append(reportMsg, msg)
		}
		ids[msg.Id] = true
	}
	p.reportedIds = ids
	return reportMsg
}

func (p *GmailPlugin) parseMessageResponse(resp *http.Response) (message, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp errorResp
		if err := decoder.Decode(&errResp); err != nil {
			return message{}, err
		}
		return message{}, &errResp
	}

	var msg message
	if err := decoder.Decode(&msg); err != nil {
		return message{}, err
	}

	return msg, nil
}

func (p *GmailPlugin) requestMessage(id, accessToken string) (*http.Response, error) {
	u, err := baseUrl.Parse("messages/" + id)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	// only request specific fields
	query.Add("fields", "snippet,threadId,id,payload/headers")
	// get the full message to get From and Subject from headers
	query.Add("format", "full")
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	return http.DefaultClient.Do(req)
}

func (p *GmailPlugin) requestMessageList(accessToken string) (*http.Response, error) {
	u, err := baseUrl.Parse("messages")
	if err != nil {
		return nil, err
	}

	query := u.Query()

	// only get unread, from the personal category that are in the inbox.
	// if we want to widen the search scope we need to add more categories
	// like: '(category:personal or category:updates or category:forums)' ...
	query.Add("q", "is:unread category:personal in:inbox")
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	return http.DefaultClient.Do(req)
}
