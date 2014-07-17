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

package main

import (
	"fmt"

	"log"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
	"launchpad.net/account-polld/plugins/gmail"
	"launchpad.net/go-dbus/v1"
)

type PostWatch struct {
	appId         plugins.ApplicationId
	notifications *[]plugins.Notification
}

const (
	SERVICENAME_GMAIL    = "google-gmail-poll"
	SERVICENAME_TWITTER  = "twitter-microblog"
	SERVICENAME_FACEBOOK = "facebook-microblog"
)

func init() {
}

func main() {
	// TODO NewAccount called here is just for playing purposes.
	postWatch := make(chan *PostWatch)

	if bus, err := dbus.Connect(dbus.SessionBus); err != nil {
		log.Fatal("Cannot connect to bus", err)
	} else {
		go postOffice(bus, postWatch)
	}

	go monitorAccounts(postWatch)

	done := make(chan bool)
	<-done
}

func monitorAccounts(postWatch chan *PostWatch) {
	mgr := make(map[uint]*AccountManager)
L:
	for data := range accounts.WatchForService(SERVICENAME_GMAIL, SERVICENAME_FACEBOOK, SERVICENAME_TWITTER) {
		if account, ok := mgr[data.AccountId]; ok {
			if data.Enabled {
				log.Printf("New account data for %d - was %#v, now is %#v", data.AccountId, account.authData, data)
				account.updateAuthData(data)
			} else {
				account.Delete()
				delete(mgr, data.AccountId)
			}
		} else if data.Enabled {
			var plugin plugins.Plugin
			switch data.ServiceName {
			case SERVICENAME_GMAIL:
				log.Println("Creating account with id", data.AccountId, "for", data.ServiceName)
				plugin = gmail.New()
			case SERVICENAME_FACEBOOK:
				// This is just stubbed until the plugin exists.
				log.Println("Unhandled account with id", data.AccountId, "for", data.ServiceName)
				continue L
			case SERVICENAME_TWITTER:
				// This is just stubbed until the plugin exists.
				log.Println("Unhandled account with id", data.AccountId, "for", data.ServiceName)
				continue L
			default:
				log.Println("Unhandled account with id", data.AccountId, "for", data.ServiceName)
				continue L
			}
			mgr[data.AccountId] = NewAccountManager(data, postWatch, plugin)
			go mgr[data.AccountId].Loop()
		}
	}
}

func postOffice(bus *dbus.Connection, postWatch chan *PostWatch) {
	for post := range postWatch {
		for _, n := range *post.notifications {
			fmt.Println("Should be dispatching", n, "to the post office using", bus.UniqueName, "for", post.appId)
		}
	}
}
