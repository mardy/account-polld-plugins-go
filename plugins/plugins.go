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

package plugins

// Plugin is an interface which the plugins will adhere to for the poll
// daemon to interact with.
//
// Poll interacts with the backend service with the means the plugin defines
// and  returns a list of Notifications to send to the Push service. If an
// error occurs and is returned the daemon can decide to throttle the service.
//
// ApplicationId returns the APP_ID of the delivery target for Post Office.
type Plugin interface {
	ApplicationId() ApplicationId
	Poll(AuthTokens) (*[]Notification, error)
}

// AuthTokens is a map with tokens the plugins are to use to make requests.
type AuthTokens map[string]interface{}

// ApplicationId represents the application id to direct posts to.
// e.g.: com.ubuntu.diaspora_diaspora or com.ubuntu.diaspora_diaspora_1.0
//
// TODO define if APP_ID can be of short form
// TODO find documentation where short APP_ID is defined (aka versionless APP_ID).
type ApplicationId string

// Notification represents the data pass over to the Post Office
// It's up to the plugin to determine if multiple Notification cards
// should be bundled together or to present them separately.
//
// The daemon can determine to throttle these depending on the
// quantity.
type Notification struct {
	Sound string `json:"sound"`
	Card  Card   `json:"card"`
}

type Card struct {
	Summary string `json:"summary"`
	Popup   bool   `json:"popup"`
	Persist bool   `json:"persist"`
}

// The constanst defined here determine the polling aggressivenes with the following criteria
// MAXIMUM: calls, health warning
// HIGH: SMS, chat message, new email
// DEFAULT: social media updates
// LOW: software updates, junk email
const (
	PRIORITY_MAXIMUM = 0
	PRIORITY_HIGH
	PRIORITY_DEFAULT
	PRIORITY_LOW
)

const (
	PLUGIN_EMAIL = 0
	PLUGIN_SOCIAL
)
