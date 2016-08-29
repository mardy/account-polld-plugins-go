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

package main

import (
	"sync"

	"log"

	"launchpad.net/account-polld/gettext"
	"launchpad.net/account-polld/plugins"
	"launchpad.net/account-polld/plugins/twitter"
	"launchpad.net/account-polld/qtcontact"
)

type AccountKey struct {
	serviceId   string
	accountId   uint
}

/* Use identifiers and API keys provided by the respective webapps which are the official
   end points for the notifications */
const (
	SERVICENAME_DEKKO     = "dekko.dekkoproject_dekko"
	SERVICENAME_GMAIL     = "com.ubuntu.developer.webapps.webapp-gmail_webapp-gmail"
	SERVICENAME_TWITTER   = "com.ubuntu.developer.webapps.webapp-twitter_webapp-twitter"
	SERVICENAME_GCALENDAR = "google-caldav"
	SERVICENAME_OCALENDAR = "owncloud-caldav"
)

var mainLoopOnce sync.Once

func init() {
	startMainLoop()
}

func startMainLoop() {
	mainLoopOnce.Do(func() {
		go qtcontact.MainLoopStart()
	})
}

func main() {
	// Initialize i18n
	gettext.SetLocale(gettext.LC_ALL, "")
	gettext.Textdomain("account-polld")
	gettext.BindTextdomain("account-polld", "/usr/share/locale")

	log.Println("Starting app")

	runner := plugins.NewPluginRunner(twitter.New())
	runner.Run()
	/*
	postWatch := make(chan *plugins.PostWatch)

	watcher := plugins.NewIpc()
	go watcher.Run()
	for {
		select {
		case data := <-watcher.C:
			log.Println("Got data, access token is ", data.AccessToken)
		case post := <-postWatch:
			log.Println("Got reply")
			watcher.PostMessages(post.batches)
		}
	}
	*/
}
