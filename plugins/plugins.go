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
// Register is the first method to be called after creating a New plugin,
// instance. upon register the plugin will check if it is Authorized and
// Installed and send it over State.
//
// Poll interacts with the backend service with the means the plugin defines
// and  returns a list of Notifications to send to the Push service. If an
// error occurs and is returned the daemon can decide to throttle the service.
//
// GetType is for future use and helps the daemon determine the polling
// frequency.
type Plugin interface {
	Register() (ApplicationId, chan (State))
	Unregister()
	Poll() (*[]Notification, error)
	//GetPriority() int
	//GetType() int
	//Notify()
}

// ApplicationId represents the application id to direct posts to.
// e.g.: com.ubuntu.diaspora_diaspora or com.ubuntu.diaspora_diaspora_1.0
//
// TODO define if APP_ID can be of short form
// TODO find documentation where short APP_ID is defined (aka versionless APP_ID).
type ApplicationId string

// Notification represents the data pass over to the Post Office
type Notification struct {
	Sound string `json:"sound"`
	Card  Card   `json:"card"`
}

type Card struct {
	Summary string `json:"summary"`
	Popup   bool   `json:"popup"`
	Persist bool   `json:"persist"`
}

// State represents a state change for a plugin. Installed is set to true when
// the package that is supposed to handle the notification is installed whilst
// Authorized being true means that at least one account is authorized to poll
type State struct {
	Installed, Authorized bool
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
