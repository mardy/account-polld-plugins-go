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
	"log"
	"time"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
)

type AccountManager struct {
	authData  accounts.AuthData
	plugin    plugins.Plugin
	interval  time.Duration
	terminate chan bool
}

const DEFAULT_INTERVAL = time.Duration(60 * time.Second)

func NewAccountManager(authData accounts.AuthData, plugin plugins.Plugin) *AccountManager {
	return &AccountManager{
		plugin:    plugin,
		authData:  authData,
		interval:  DEFAULT_INTERVAL,
		terminate: make(chan bool),
	}
}

func (a *AccountManager) Delete() {
	a.terminate <- true
}

func (a *AccountManager) Loop(postWatch chan *PostWatch) {
	defer close(a.terminate)
L:
	for {
		log.Println("Polling set to", a.interval)
		select {
		case <-time.After(a.interval):
			if n, err := a.plugin.Poll(&a.authData); err != nil {
				log.Print("Error while polling ", a.authData.AccountId, ": ", err)
				// penalizing the next poll
				a.interval += DEFAULT_INTERVAL
				continue
			} else if n != nil {
				// on success we reset the timeout to the default interval
				a.interval = DEFAULT_INTERVAL
				postWatch <- &PostWatch{notifications: n, appId: a.plugin.ApplicationId()}
			}
		case <-a.terminate:
			break L
		}
	}
	log.Printf("Ending poll loop for account %d", a.authData.AccountId)
}
