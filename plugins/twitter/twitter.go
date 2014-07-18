/*
 Copyright 2014 Canonical Ltd.
 Authors: James Henstridge <james.henstridge@canonical.com>

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

package twitter

import (
	"fmt"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/garyburd/go-oauth/oauth"
	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
)

var baseUrl, _ = url.Parse("https://api.twitter.com/1.1/")

type twitterPlugin struct {
	lastMentionId int64
	lastDirectMessageId int64
}

func New() plugins.Plugin {
	return &twitterPlugin{}
}

func (p *twitterPlugin) ApplicationId() plugins.ApplicationId {
	return "com.ubuntu.developer.webapps.webapp-twitter_webapp-twitter"
}

func (p *twitterPlugin) request(authData *accounts.AuthData, path string) (*http.Response, error) {
	// Resolve path relative to Graph API base URL, and add access token
	u, err := baseUrl.Parse(path)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	u.RawQuery = ""

	client := oauth.Client{
		Credentials: oauth.Credentials{
			Token: authData.ClientId,
			Secret: authData.ClientSecret,
		},
	}
	token := &oauth.Credentials{
		Token: authData.AccessToken,
		Secret: authData.TokenSecret,
	}
	return client.Get(http.DefaultClient, token, u.String(), query)
}

func (p *twitterPlugin) parseStatuses(resp *http.Response) ([]plugins.PushMessage, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var result TwitterError
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}
		return nil, &result
	}

	var statuses []status
	if err := decoder.Decode(&statuses); err != nil {
		return nil, err
	}
	pushMsg := []plugins.PushMessage{}
	latestStatus := p.lastMentionId
	for _, s := range statuses {
		pushMsg = append(pushMsg, plugins.PushMessage{
			Notification: plugins.Notification{
				Card: &plugins.Card{
					Summary: fmt.Sprintf("Mention from @%s", s.User.ScreenName),
					Body: s.Text,
				},
			},
		})
		if s.Id > latestStatus {
			latestStatus = s.Id
		}
	}
	p.lastMentionId = latestStatus
	return pushMsg, nil
}

func (p *twitterPlugin) parseDirectMessages(resp *http.Response) ([]plugins.PushMessage, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var result TwitterError
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}
		return nil, &result
	}

	var dms []directMessage
	if err := decoder.Decode(&dms); err != nil {
		return nil, err
	}
	pushMsg := []plugins.PushMessage{}
	latestDM := p.lastDirectMessageId
	for _, m := range dms {
		pushMsg = append(pushMsg, plugins.PushMessage{
			Notification: plugins.Notification{
				Card: &plugins.Card{
					Summary: fmt.Sprintf("Direct message from @%s", m.Sender.ScreenName),
					Body: m.Text,
				},
			},
		})
		if m.Id > latestDM {
			latestDM = m.Id
		}
	}
	p.lastDirectMessageId = latestDM
	return pushMsg, nil
}

func (p *twitterPlugin) Poll(authData *accounts.AuthData) (messages []plugins.PushMessage, err error) {
	url := "statuses/mentions_timeline.json"
	if p.lastMentionId > 0 {
		url = fmt.Sprintf("%s?since_id=%d", url, p.lastMentionId)
	}
	resp, err := p.request(authData, url)
	if err != nil {
		return
	}
	messages, err = p.parseStatuses(resp)
	if err != nil {
		return
	}

	url = "direct_messages.json"
	if p.lastDirectMessageId > 0 {
		url = fmt.Sprintf("%s?since_id=%d", url, p.lastDirectMessageId)
	}
	resp, err = p.request(authData, url)
	if err != nil {
		return
	}
	dms, err := p.parseDirectMessages(resp)
	if err != nil {
		return
	}
	messages = append(messages, dms...)
	return
}

// Status format is described here:
// https://dev.twitter.com/docs/api/1.1/get/statuses/mentions_timeline
type status struct {
	Id int64 `json:"id"`
	CreatedAt string `json:"created_at"`
	User user `json:"user"`
	Text string `json:"text"`
}

// Direct message format is described here:
// https://dev.twitter.com/docs/api/1.1/get/direct_messages
type directMessage struct {
	Id int64 `json:"id"`
	CreatedAt string `json:"created_at"`
	Sender user `json:"sender"`
	Recipient user `json:"recipient"`
	Text string `json:"text"`
}

	type user struct {
		Id int64 `json:"id"`
	ScreenName string `json:"screen_name"`
	Name string `json:"name"`
	Image string `json:"profile_image_url"`

}

// The error response format is described here:
// https://dev.twitter.com/docs/error-codes-responses
type TwitterError struct {
	Errors []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func (err *TwitterError) Error() string {
	messages := make([]string, len(err.Errors))
	for i := range(err.Errors) {
		messages[i] = err.Errors[i].Message
	}
	return strings.Join(messages, "\n")
}
