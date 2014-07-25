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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
	"launchpad.net/account-polld/plugins/twitter/oauth" // "github.com/garyburd/go-oauth/oauth"
)

var baseUrl, _ = url.Parse("https://api.twitter.com/1.1/")

const twitterIcon = "/usr/share/click/preinstalled/.click/users/@all/com.ubuntu.developer.webapps.webapp-twitter/twitter.png"

const (
	maxIndividualStatuses               = 2
	consolidatedStatusIndexStart        = maxIndividualStatuses
	maxIndividualDirectMessages         = 2
	consolidatedDirectMessageIndexStart = maxIndividualDirectMessages
)

type twitterPlugin struct {
	lastMentionId       int64
	lastDirectMessageId int64
}

func New() plugins.Plugin {
	return &twitterPlugin{}
}

func (p *twitterPlugin) ApplicationId() plugins.ApplicationId {
	return "com.ubuntu.developer.webapps.webapp-twitter_webapp-twitter"
}

func (p *twitterPlugin) request(authData *accounts.AuthData, path string) (*http.Response, error) {
	// Resolve path relative to API base URL.
	u, err := baseUrl.Parse(path)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	u.RawQuery = ""

	client := oauth.Client{
		Credentials: oauth.Credentials{
			Token:  authData.ClientId,
			Secret: authData.ClientSecret,
		},
	}
	token := &oauth.Credentials{
		Token:  authData.AccessToken,
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

	sort.Sort(sort.Reverse(byStatusId(statuses)))
	if len(statuses) < 1 {
		return nil, nil
	}
	p.lastMentionId = statuses[0].Id

	pushMsg := []plugins.PushMessage{}
	for _, s := range statuses {
		pushMsg = append(pushMsg, plugins.PushMessage{
			Notification: plugins.Notification{
				Card: &plugins.Card{
					Summary: fmt.Sprintf("@%s mentioned you", s.User.ScreenName),
					Body:    s.Text,
					Actions: []string{fmt.Sprintf("http://mobile.twitter.com/%s/statuses/%d", s.User.ScreenName, s.Id)},
					Icon:    twitterIcon,
					Persist: true,
					Popup:   true,
				},
				Sound:   plugins.DefaultSound(),
				Vibrate: plugins.DefaultVibration(),
			},
		})
		if len(pushMsg) == maxIndividualStatuses {
			break
		}
	}
	// Now we consolidate the remaining statuses
	if len(statuses) > len(pushMsg) {
		var screennames []string
		for _, s := range statuses[consolidatedStatusIndexStart:] {
			screennames = append(screennames, s.User.ScreenName)
		}
		pushMsg = append(pushMsg, plugins.PushMessage{
			Notification: plugins.Notification{
				Card: &plugins.Card{
					Summary: "Multiple more mentions",
					Body:    fmt.Sprintf("From %s", strings.Join(screennames, ", ")),
					Actions: []string{"http://mobile.twitter.com/i/connect"},
					Icon:    twitterIcon,
					Persist: true,
					Popup:   true,
				},
				Sound:   plugins.DefaultSound(),
				Vibrate: plugins.DefaultVibration(),
			},
		})
	}
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

	sort.Sort(sort.Reverse(byDMId(dms)))
	if len(dms) < 1 {
		return nil, nil
	}
	p.lastDirectMessageId = dms[0].Id

	pushMsg := []plugins.PushMessage{}
	for _, m := range dms {
		pushMsg = append(pushMsg, plugins.PushMessage{
			Notification: plugins.Notification{
				Card: &plugins.Card{
					Summary: fmt.Sprintf("@%s sent you a DM", m.Sender.ScreenName),
					Body:    m.Text,
					Actions: []string{fmt.Sprintf("http://mobile.twitter.com/%s/messages", m.Sender.ScreenName)},
					Icon:    twitterIcon,
					Persist: true,
					Popup:   true,
				},
				Sound:   plugins.DefaultSound(),
				Vibrate: plugins.DefaultVibration(),
			},
		})
		if len(pushMsg) == maxIndividualDirectMessages {
			break
		}
	}
	// Now we consolidate the remaining messages
	if len(dms) > len(pushMsg) {
		var senders []string
		for _, m := range dms[consolidatedDirectMessageIndexStart:] {
			senders = append(senders, m.Sender.ScreenName)
		}
		pushMsg = append(pushMsg, plugins.PushMessage{
			Notification: plugins.Notification{
				Card: &plugins.Card{
					Summary: "Multiple direct messages available",
					Body:    fmt.Sprintf("From %s", strings.Join(senders, ", ")),
					Actions: []string{"http://mobile.twitter.com/messages"},
					Icon:    twitterIcon,
					Persist: true,
					Popup:   true,
				},
				Sound:   plugins.DefaultSound(),
				Vibrate: plugins.DefaultVibration(),
			},
		})
	}
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
	Id        int64  `json:"id"`
	CreatedAt string `json:"created_at"`
	User      user   `json:"user"`
	Text      string `json:"text"`
}

// ByStatusId implements sort.Interface for []status based on
// the Id field.
type byStatusId []status

func (s byStatusId) Len() int           { return len(s) }
func (s byStatusId) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byStatusId) Less(i, j int) bool { return s[i].Id < s[j].Id }

// Direct message format is described here:
// https://dev.twitter.com/docs/api/1.1/get/direct_messages
type directMessage struct {
	Id        int64  `json:"id"`
	CreatedAt string `json:"created_at"`
	Sender    user   `json:"sender"`
	Recipient user   `json:"recipient"`
	Text      string `json:"text"`
}

// ByStatusId implements sort.Interface for []status based on
// the Id field.
type byDMId []directMessage

func (s byDMId) Len() int           { return len(s) }
func (s byDMId) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byDMId) Less(i, j int) bool { return s[i].Id < s[j].Id }

type user struct {
	Id         int64  `json:"id"`
	ScreenName string `json:"screen_name"`
	Name       string `json:"name"`
	Image      string `json:"profile_image_url"`
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
	for i := range err.Errors {
		messages[i] = err.Errors[i].Message
	}
	return strings.Join(messages, "\n")
}
