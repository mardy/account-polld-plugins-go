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
	"sort"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
)

const APP_ID = "com.ubuntu.developer.webapps.webapp-gmail_webapp-gmail"

var baseUrl, _ = url.Parse("https://www.googleapis.com/gmail/v1/users/me/")

type GmailPlugin struct {
	// reportedIds holds the messages that have already been notified. This
	// approach is taken against timestamps as it avoids needing to call
	// get on the message.
	//
	// The potential missed notification is when the user manually marks a
	// message as unread that is already in this list (which could be
	// considered a good thing).
	//
	// TODO determine if persisting the list to avoid renotification on reboot.
	// TODO clean the list with messages that are read.
	reportedIds []string
}

func New() *GmailPlugin {
	return &GmailPlugin{}
}

func (p *GmailPlugin) ApplicationId() plugins.ApplicationId {
	return plugins.ApplicationId(APP_ID)
}

func (p *GmailPlugin) Poll(authData *accounts.AuthData) ([]plugins.PushMessage, error) {
	resp, err := p.requestMessageList(authData.AccessToken)
	if err != nil {
		return nil, err
	}
	messages, err := p.parseMessageListResponse(resp)
	if err != nil {
		return nil, err
	}
	for i := range messages {
		// Don't download message payload for previously reported messages
		if p.reported(messages[i].Id) {
			continue
		}
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
	sort.Strings(p.reportedIds)
	if i := sort.SearchStrings(p.reportedIds, id); i < len(p.reportedIds) {
		return p.reportedIds[i] == id
	}
	return false
}

func (p *GmailPlugin) createNotifications(messages []message) ([]plugins.PushMessage, error) {
	pushMsgMap := make(pushes)

	for _, msg := range messages {
		// Don't report previously reported messages
		if p.reported(msg.Id) {
			continue
		}
		p.reportedIds = append(p.reportedIds, msg.Id)
		hdr := msg.Payload.mapHeaders()
		if _, ok := pushMsgMap[msg.ThreadId]; ok {
			pushMsgMap[msg.ThreadId].Notification.Card.Summary += fmt.Sprintf(", %s", hdr[hdr_FROM])
		} else {
			pushMsgMap[msg.ThreadId] = plugins.PushMessage{
				Notification: plugins.Notification{
					Card: &plugins.Card{
						Summary: fmt.Sprintf("Message \"%s\" from %s", hdr[hdr_SUBJECT], hdr[hdr_FROM]),
						Body:    msg.Snippet,
						// TODO this is a placeholder, Actions aren't fully defined yet and opening
						// multiple inboxes has issues.
						Actions: []string{"Open", "https://mail.google.com/mail/u/0/?pli=1#inbox/" + msg.ThreadId},
						Popup:   true,
						Persist: true,
					},
				},
			}
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
		return nil, &errResp
	}

	var messages messageList
	if err := decoder.Decode(&messages); err != nil {
		return nil, err
	}

	return messages.Messages, nil
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
	query.Add("fields", "snippet,threadId,id,payload")
	// get the full message to get From and Subject from headers
	query.Add("format", "full")
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	return client.Do(req)
}

func (p *GmailPlugin) requestMessageList(accessToken string) (*http.Response, error) {
	u, err := baseUrl.Parse("messages")
	if err != nil {
		return nil, err
	}

	query := u.Query()

	// only get unread
	query.Add("q", "is:unread")
	// from the INBOX
	query.Add("labelIds", "INBOX")
	// from the Personal category.
	query.Add("labelIds", "CATEGORY_PERSONAL")
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	return client.Do(req)
}
