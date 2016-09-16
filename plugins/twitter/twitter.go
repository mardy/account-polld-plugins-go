/*
 Copyright 2014 Canonical Ltd.

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
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"launchpad.net/account-polld/gettext"
	"launchpad.net/account-polld/plugins"
	"launchpad.net/account-polld/plugins/twitter/oauth" // "github.com/garyburd/go-oauth/oauth"
)

var baseUrl, _ = url.Parse("https://api.twitter.com/1.1/")

const (
	maxIndividualStatuses               = 2
	consolidatedStatusIndexStart        = maxIndividualStatuses
	maxIndividualDirectMessages         = 2
	consolidatedDirectMessageIndexStart = maxIndividualDirectMessages
	twitterDispatchUrlBase              = "https://mobile.twitter.com"
	pluginName                          = "twitter"
)

type twitterConfig struct {
	LastMentionId       int64 `json:"lastMentionId"`
	LastDirectMessageId int64 `json:"lastDirectMessageId"`
}

type twitterPlugin struct {
	config twitterConfig
}

func New() plugins.Plugin {
	return &twitterPlugin{}
}

func (p *twitterPlugin) ApplicationId() plugins.ApplicationId {
	return "com.ubuntu.developer.webapps.webapp-twitter_webapp-twitter"
}

func (p *twitterPlugin) request(authData *plugins.AuthData, path string) (*http.Response, error) {
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

func (p *twitterPlugin) parseStatuses(resp *http.Response) (*plugins.PushMessageBatch, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var result TwitterError
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}
		// Error code 89 is used for invalid or expired tokens.
		for _, e := range result.Errors {
			if e.Code == 89 {
				return nil, plugins.ErrTokenExpired
			}
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
	p.config.LastMentionId = statuses[0].Id

	pushMsg := make([]*plugins.PushMessage, len(statuses))
	for i, s := range statuses {
		// TRANSLATORS: The first %s refers to the twitter user's Name, the second %s to the username.
		summary := fmt.Sprintf(gettext.Gettext("%s. @%s"), s.User.Name, s.User.ScreenName)
		action := fmt.Sprintf("%s/%s/statuses/%d", twitterDispatchUrlBase, s.User.ScreenName, s.Id)
		epoch := toEpoch(s.CreatedAt)
		pushMsg[i] = plugins.NewStandardPushMessage(summary, s.Text, action, s.User.Image, epoch)
	}
	return &plugins.PushMessageBatch{
		Messages:        pushMsg,
		Limit:           maxIndividualStatuses,
		OverflowHandler: p.consolidateStatuses,
		Tag:             "status",
	}, nil
}

func (p *twitterPlugin) consolidateStatuses(pushMsg []*plugins.PushMessage) *plugins.PushMessage {
	screennames := make([]string, len(pushMsg))
	for i, m := range pushMsg {
		screennames[i] = m.Notification.Card.Summary
	}
	// TRANSLATORS: This represents a notification summary about more twitter mentions available
	summary := gettext.Gettext("Multiple more mentions")
	// TRANSLATORS: This represents a notification body with the comma separated twitter usernames
	body := fmt.Sprintf(gettext.Gettext("From %s"), strings.Join(screennames, ", "))
	action := fmt.Sprintf("%s/i/connect", twitterDispatchUrlBase)
	epoch := time.Now().Unix()

	return plugins.NewStandardPushMessage(summary, body, action, "", epoch)
}

func (p *twitterPlugin) parseDirectMessages(resp *http.Response) (*plugins.PushMessageBatch, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var result TwitterError
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}
		// Error code 89 is used for invalid or expired tokens.
		for _, e := range result.Errors {
			if e.Code == 89 {
				return nil, plugins.ErrTokenExpired
			}
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
	p.config.LastDirectMessageId = dms[0].Id

	pushMsg := make([]*plugins.PushMessage, len(dms))
	for i, m := range dms {
		// TRANSLATORS: The first %s refers to the twitter user's Name, the second %s to the username.
		summary := fmt.Sprintf(gettext.Gettext("%s. @%s"), m.Sender.Name, m.Sender.ScreenName)
		action := fmt.Sprintf("%s/%s/messages", twitterDispatchUrlBase, m.Sender.ScreenName)
		epoch := toEpoch(m.CreatedAt)
		pushMsg[i] = plugins.NewStandardPushMessage(summary, m.Text, action, m.Sender.Image, epoch)
	}

	return &plugins.PushMessageBatch{
		Messages:        pushMsg,
		Limit:           maxIndividualDirectMessages,
		OverflowHandler: p.consolidateDirectMessages,
		Tag:             "direct-message",
	}, nil
}

func (p *twitterPlugin) consolidateDirectMessages(pushMsg []*plugins.PushMessage) *plugins.PushMessage {
	senders := make([]string, len(pushMsg))
	for i, m := range pushMsg {
		senders[i] = m.Notification.Card.Summary
	}
	// TRANSLATORS: This represents a notification summary about more twitter direct messages available
	summary := gettext.Gettext("Multiple direct messages available")
	// TRANSLATORS: This represents a notification body with the comma separated twitter usernames
	body := fmt.Sprintf(gettext.Gettext("From %s"), strings.Join(senders, ", "))
	action := fmt.Sprintf("%s/messages", twitterDispatchUrlBase)
	epoch := time.Now().Unix()

	return plugins.NewStandardPushMessage(summary, body, action, "", epoch)
}

func (p *twitterPlugin) loadPersistentData(accountId uint) error {
	err := plugins.FromPersist(pluginName, accountId, &p.config)
	return err
}

func (p *twitterPlugin) savePersistentData(accountId uint) error {
	err := plugins.Persist(pluginName, accountId, p.config)
	if err != nil {
		log.Print("twitter plugin ", accountId, ": failed to save state: ", err)
	}
	return err
}

func (p *twitterPlugin) Poll(authData *plugins.AuthData) (batches []*plugins.PushMessageBatch, err error) {
	defer p.savePersistentData(authData.AccountId)
	p.loadPersistentData(authData.AccountId)

	url := "statuses/mentions_timeline.json"
	notifyMentions := false
	if p.config.LastMentionId > 0 {
		url = fmt.Sprintf("%s?since_id=%d", url, p.config.LastMentionId)
		notifyMentions = true
	}
	resp, err := p.request(authData, url)
	if err != nil {
		return
	}
	statuses, err := p.parseStatuses(resp)
	if err != nil {
		return
	}

	url = "direct_messages.json"
	notifyMessages := false
	if p.config.LastDirectMessageId > 0 {
		url = fmt.Sprintf("%s?since_id=%d", url, p.config.LastDirectMessageId)
		notifyMessages = true
	}
	resp, err = p.request(authData, url)
	if err != nil {
		return
	}
	dms, err := p.parseDirectMessages(resp)
	if err != nil {
		return
	}
	// Don't send any notifications if this is the very first time we run:
	// otherwise we would emit notifications for all the old messages too
	if notifyMentions {
		if statuses != nil && len(statuses.Messages) > 0 {
			batches = append(batches, statuses)
		}
	}
	if notifyMessages {
		if dms != nil && len(dms.Messages) > 0 {
			batches = append(batches, dms)
		}
	}
	return
}

func toEpoch(timestamp string) int64 {
	if t, err := time.Parse(time.RubyDate, timestamp); err == nil {
		return t.Unix()
	}
	return time.Now().Unix()
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
