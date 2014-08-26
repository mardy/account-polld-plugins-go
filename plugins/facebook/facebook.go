/*
 Copyright 2014 Canonical Ltd.
 Authors: James Henstridge <james.henstridge@canonical.com>
          Sergio Schvezov <sergio.schvezov@canonical.com>

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

package facebook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"log"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/gettext"
	"launchpad.net/account-polld/plugins"
)

const (
	facebookTime                        = "2006-01-02T15:04:05-0700"
	maxIndividualNotifications          = 2
	consolidatedNotificationsIndexStart = maxIndividualNotifications
	maxIndividualThreads                = 2
	consolidatedThreadsIndexStart       = maxIndividualThreads
	pluginName                          = "facebook"
)

var baseUrl, _ = url.Parse("https://graph.facebook.com/v2.0/")

type timeStamp string

func (state fbState) persist(accountId uint) (err error) {
	err = plugins.Persist(pluginName, accountId, state)
	if err != nil {
		log.Print("facebook plugin", accountId, ": failed to save state: ", err)
	}
	return nil
}

func stateFromStorage(accountId uint) (state fbState, err error) {
	err = plugins.FromPersist(pluginName, accountId, &state)
	if err != nil {
		return state, err
	}
	if _, err := time.Parse(facebookTime, string(state.lastUpdate)); err == nil {
		return state, err
	}
	if _, err := time.Parse(facebookTime, string(state.lastInboxUpdate)); err == nil {
		return state, err
	}
	return state, nil
}

type fbState struct {
	lastUpdate      timeStamp `json:"lastNotificationUpdate"`
	lastInboxUpdate timeStamp `json:"lastInboxUpdate"`
}

type fbPlugin struct {
	state     fbState
	accountId uint
}

func New(accountId uint) plugins.Plugin {
	state, err := stateFromStorage(accountId)
	if err != nil {
		log.Print("facebook plugin ", accountId, ": cannot load previous state from storage: ", err)
	} else {
		log.Print("facebook plugin ", accountId, ": last state loaded from storage")
	}
	return &fbPlugin{state: state, accountId: accountId}
}

func (p *fbPlugin) ApplicationId() plugins.ApplicationId {
	return "com.ubuntu.developer.webapps.webapp-facebook_webapp-facebook"
}

func (p *fbPlugin) request(authData *accounts.AuthData, path string) (*http.Response, error) {
	// Resolve path relative to Graph API base URL, and add access token
	u, err := baseUrl.Parse(path)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	query.Add("access_token", authData.AccessToken)
	u.RawQuery = query.Encode()

	return http.Get(u.String())
}

func (p *fbPlugin) checkError(resp *http.Response, decoder *json.Decoder) error {
	if resp.StatusCode != http.StatusOK {
		var result errorDoc
		if err := decoder.Decode(&result); err != nil {
			return err
		}
		if result.Error.Code == 190 {
			return plugins.ErrTokenExpired
		}
		return &result.Error
	}
	return nil
}

func (p *fbPlugin) parseResponse(resp *http.Response) ([]plugins.PushMessage, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err := p.checkError(resp, decoder)
	if err != nil {
		return nil, err
	}

	// TODO: Follow the "paging.next" link if we get more than one
	// page full of notifications.  The default limit seems to be
	// 5000 though, which we are unlikely to hit, since
	// notifications are deleted once read.
	// TODO filter out of date messages before operating
	var result notificationDoc
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}

	var validNotifications []Notification
	latestUpdate := p.state.lastUpdate
	for i, n := range result.Data {
		if !n.isValid(p.state.lastUpdate) {
			log.Println("facebook plugin: skipping notification", n.Id, "as", n.UpdatedTime, "!=", p.state.lastUpdate, "or", n.Unread)
		} else {
			log.Println("facebook plugin: valid notification", n.Id, "dated:", n.UpdatedTime, "and read status:", n.Unread)
			validNotifications = append(validNotifications, &result.Data[i]) // get the actual reference, not the copy
			if n.UpdatedTime > latestUpdate {
				latestUpdate = n.UpdatedTime
			}
		}
	}
	p.state.lastUpdate = latestUpdate
	p.state.persist(p.accountId)

	pushMsg := []plugins.PushMessage{}
	for _, n := range validNotifications {
		msg := n.buildPushMessage()
		pushMsg = append(pushMsg, *msg)
		if len(pushMsg) == maxIndividualNotifications {
			break
		}
	}

	// Now we consolidate the remaining statuses
	if len(validNotifications) > len(pushMsg) && len(validNotifications) >= consolidatedNotificationsIndexStart {
		usernamesMap := result.getConsolidatedMessagesUsernames(consolidatedNotificationsIndexStart)
		usernames := []string{}
		for k, _ := range usernamesMap {
			usernames = append(usernames, k)
			// we don't too many usernames listed, this is a hard number
			if len(usernames) > 10 {
				usernames = append(usernames, "...")
				break
			}
		}
		if len(usernames) > 0 {
			consolidatedMsg := result.getConsolidatedMessage(usernames)
			pushMsg = append(pushMsg, *consolidatedMsg)
		}
	}

	return pushMsg, nil
}

func (p *fbPlugin) parseInboxResponse(resp *http.Response) ([]plugins.PushMessage, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err := p.checkError(resp, decoder)
	if err != nil {
		return nil, err
	}

	// TODO: Follow the "paging.next" link, need to check what the default limit is, etc.
	// TODO filter out of date messages before operating
	var result inboxDoc
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}

	var validThreads []Notification
	latestUpdate := p.state.lastUpdate
	for i, t := range result.Data {
		if !t.isValid(p.state.lastInboxUpdate) {
			// already seen, move on
			log.Println("facebook plugin: skipping seen thread:", t)
		} else {
			// process this thread
			log.Println("facebook plugin: valid thread", t.Id, "dated:", t.UpdatedTime, "and read status:", t.Unread)
			validThreads = append(validThreads, &result.Data[i]) // get the actual reference, not the copy
			if t.UpdatedTime > latestUpdate {
				latestUpdate = t.UpdatedTime
			}
		}
	}
	p.state.lastInboxUpdate = latestUpdate
	p.state.persist(p.accountId)

	pushMsg := []plugins.PushMessage{}
	fmt.Println("valid Threads", validThreads)
	for _, t := range validThreads {
		msg := t.buildPushMessage()
		pushMsg = append(pushMsg, *msg)
		if len(pushMsg) == maxIndividualThreads {
			break
		}
	}

	// Now we consolidate the remaining messages
	if len(validThreads) > len(pushMsg) && len(validThreads) >= consolidatedThreadsIndexStart {
		usernamesMap := result.getConsolidatedMessagesUsernames(consolidatedThreadsIndexStart)
		usernames := []string{}
		for _, v := range usernamesMap {
			usernames = append(usernames, v)
			// we don't too many usernames listed, this is a hard number
			if len(usernames) > 10 {
				usernames = append(usernames, "...")
				break
			}
		}
		if len(usernames) > 0 {
			consolidatedMsg := result.getConsolidatedMessage(usernames)
			pushMsg = append(pushMsg, *consolidatedMsg)
		}
	}
	return pushMsg, nil
}

func (p *fbPlugin) Poll(authData *accounts.AuthData) ([]plugins.PushMessage, error) {
	// This envvar check is to ease testing.
	if token := os.Getenv("ACCOUNT_POLLD_TOKEN_FACEBOOK"); token != "" {
		authData.AccessToken = token
	}
	resp, err := p.request(authData, "me/notifications")
	if err != nil {
		return nil, err
	}
	notifications, err := p.parseResponse(resp)
	resp, err = p.request(authData, "me/inbox?fields=unread,unseen,comments.limit(1)")
	inbox, err := p.parseInboxResponse(resp)
	notifSize := len(notifications)
	inboxSize := len(inbox)
	var messages = make([]plugins.PushMessage, notifSize+inboxSize)
	for i, v := range notifications {
		messages[i] = v
	}
	for i, v := range inbox {
		messages[notifSize+i] = v
	}
	return messages, err
}

func toEpoch(stamp timeStamp) int64 {
	if t, err := time.Parse(facebookTime, string(stamp)); err == nil {
		return t.Unix()
	}
	return time.Now().Unix()
}

// The notifications response format is described here:
// https://developers.facebook.com/docs/graph-api/reference/v2.0/user/notifications/
type notificationDoc struct {
	Data   []notification `json:"data"`
	Paging struct {
		Previous string `json:"previous"`
		Next     string `json:"next"`
	} `json:"paging"`
}

type notification struct {
	Id          string    `json:"id"`
	From        object    `json:"from"`
	To          object    `json:"to"`
	CreatedTime timeStamp `json:"created_time"`
	UpdatedTime timeStamp `json:"updated_time"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Application object    `json:"application"`
	Unread      int       `json:"unread"`
	Object      object    `json:"object"`
}

func picture(msgId string, id string) string {
	u, err := baseUrl.Parse(fmt.Sprintf("%s/picture", id))
	if err != nil {
		log.Println("facebook plugin: cannot get picture for", msgId)
		return ""
	}
	query := u.Query()
	query.Add("redirect", "true")
	u.RawQuery = query.Encode()
	return u.String()
}

type object struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// The error response format is described here:
// https://developers.facebook.com/docs/graph-api/using-graph-api/v2.0#errors
type errorDoc struct {
	Error GraphError `json:"error"`
}

type GraphError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    int    `json:"code"`
	Subcode int    `json:"error_subcode"`
}

func (err *GraphError) Error() string {
	return err.Message
}

// The inbox response format is described here:
// https://developers.facebook.com/docs/graph-api/reference/v2.0/user/inbox
type inboxDoc struct {
	Data   []thread `json:"data"`
	Paging struct {
		Previous string `json:"previous"`
		Next     string `json:"next"`
	} `json:"paging"`
	Summary summary `json:"summary"`
}

type summary struct {
	UnseenCount int       `json:"unseen_count"`
	UnreadCount int       `json:"unread_count"`
	UpdatedTime timeStamp `json:"updated_time"`
}

type thread struct {
	Id          string    `json:"id"`
	Comments    comments  `json:"comments"`
	To          []object  `json:"to"`
	Unread      int       `json:"unread"`
	Unseen      int       `json:"unseen"`
	UpdatedTime timeStamp `json:"updated_time"`
	Paging      struct {
		Previous string `json:"previous"`
		Next     string `json:"next"`
	} `json:"paging"`
}

type comments struct {
	Data   []message `json:"data"`
	Paging struct {
		Previous string `json:"previous"`
		Next     string `json:"next"`
	} `json:"paging"`
}

type message struct {
	CreatedTime timeStamp `json:"created_time"`
	From        object    `json:"from"`
	Id          string    `json:"id"`
	Message     string    `json:message`
}

func (doc *inboxDoc) getConsolidatedMessagesUsernames(idxStart int) map[string]string {
	usernamesMap := make(map[string]string)
	for _, t := range doc.Data[idxStart-1:] {
		message := t.Comments.Data[0]
		userId := message.From.Id
		if _, ok := usernamesMap[userId]; !ok {
			username := message.From.Name
			usernamesMap[userId] = username
		}
	}
	return usernamesMap
}

func (doc *inboxDoc) getConsolidatedMessage(usernames []string) *plugins.PushMessage {
	// TRANSLATORS: This represents a message summary about more facebook messages
	summary := gettext.Gettext("Multiple more messages")
	// TRANSLATORS: This represents a message body with the comma separated facebook usernames
	body := fmt.Sprintf(gettext.Gettext("From %s"), strings.Join(usernames, ", "))
	action := "https://m.facebook.com/messages"
	epoch := time.Now().Unix()
	return plugins.NewStandardPushMessage(summary, body, action, "", epoch)
}

func (t *thread) buildPushMessage() *plugins.PushMessage {
	link := "https://www.facebook.com/messages?action=recent-messages"
	epoch := toEpoch(t.UpdatedTime)
	// get the single message we fetch
	message := t.Comments.Data[0]
	return plugins.NewStandardPushMessage(message.From.Name, message.Message, link, picture(t.Id, message.From.Id), epoch)
}

func (t *thread) isValid(tStamp timeStamp) bool {
	return !(t.UpdatedTime <= tStamp || t.Unread == 0 || t.Unseen == 0)
}

func (doc *notificationDoc) getConsolidatedMessagesUsernames(idxStart int) map[string]string {
	usernamesMap := make(map[string]string)
	for _, n := range doc.Data[idxStart:] {
		if _, ok := usernamesMap[n.From.Id]; !ok {
			usernamesMap[n.From.Id] = n.From.Name
		}
	}
	return usernamesMap
}

func (doc *notificationDoc) getConsolidatedMessage(usernames []string) *plugins.PushMessage {
	// TRANSLATORS: This represents a notification summary about more facebook notifications
	summary := gettext.Gettext("Multiple more notifications")
	// TRANSLATORS: This represents a notification body with the comma separated facebook usernames
	body := fmt.Sprintf(gettext.Gettext("From %s"), strings.Join(usernames, ", "))
	action := "https://m.facebook.com"
	epoch := time.Now().Unix()
	return plugins.NewStandardPushMessage(summary, body, action, "", epoch)
}

func (n *notification) buildPushMessage() *plugins.PushMessage {
	epoch := toEpoch(n.UpdatedTime)
	return plugins.NewStandardPushMessage(n.From.Name, n.Title, n.Link, picture(n.Id, n.From.Id), epoch)
}

func (n *notification) isValid(tStamp timeStamp) bool {
	return n.UpdatedTime > tStamp && n.Unread >= 1
}

type Document interface {
	getConsolidatedMessage([]string) *plugins.PushMessage
	getConsolidatedMessagesUsernames(int) map[string]string
}

type Notification interface {
	buildPushMessage() *plugins.PushMessage
	isValid(timeStamp) bool
}

var _ Notification = (*thread)(nil)
var _ Notification = (*notification)(nil)
