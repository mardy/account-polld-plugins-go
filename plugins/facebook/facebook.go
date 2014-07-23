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

package facebook

import (
	"encoding/json"
	"net/http"
	"net/url"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
)

var baseUrl, _ = url.Parse("https://graph.facebook.com/v2.0/")

type fbPlugin struct {
	lastUpdate string
}

func New() plugins.Plugin {
	return &fbPlugin{}
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

func (p *fbPlugin) parseResponse(resp *http.Response) ([]plugins.PushMessage, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var result errorDoc
		if err := decoder.Decode(&result); err != nil {
			return nil, err
		}
		if result.Error.Code == 190 {
			return nil, plugins.ErrTokenExpired
		}
		return nil, &result.Error
	}

	// TODO: Follow the "paging.next" link if we get more than one
	// page full of notifications.  The default limit seems to be
	// 5000 though, which we are unlikely to hit, since
	// notifications are deleted once read.
	var result notificationDoc
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	pushMsg := []plugins.PushMessage{}
	latestUpdate := ""
	for _, n := range result.Data {
		if n.UpdatedTime <= p.lastUpdate {
			continue
		}
		pushMsg = append(pushMsg, plugins.PushMessage{
			Notification: plugins.Notification{
				Card: &plugins.Card{
					Summary: n.Title,
				},
			},
		})
		if n.UpdatedTime > latestUpdate {
			latestUpdate = n.UpdatedTime
		}
	}
	p.lastUpdate = latestUpdate
	return pushMsg, nil
}

func (p *fbPlugin) Poll(authData *accounts.AuthData) ([]plugins.PushMessage, error) {
	resp, err := p.request(authData, "me/notifications")
	if err != nil {
		return nil, err
	}
	return p.parseResponse(resp)
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
	Id          string `json:"id"`
	From        object `json:"from"`
	To          object `json:"to"`
	CreatedTime string `json:"created_time"`
	UpdatedTime string `json:"updated_time"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Application object `json:"application"`
	Unread      int    `json:"unread"`
	Object      object `json:"object"`
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
