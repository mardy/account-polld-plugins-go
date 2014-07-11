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

type Account struct {
	authData     *accounts.AuthData
	plugin       plugins.Plugin
	interval     time.Duration
	installWatch chan bool
}

const DEFAULT_INTERVAL = time.Duration(10 * time.Second)

func NewAccount(authData *accounts.AuthData, plugin plugins.Plugin) *Account {
	return &Account{
		plugin:       plugin,
		authData:     authData,
		interval:     DEFAULT_INTERVAL,
		installWatch: make(chan bool),
	}
}

func (a *Account) Delete() {
	close(a.installWatch)
}

func (a *Account) Loop(postWatch chan *PostWatch) {
	log.Println("Polling set to", a.interval)
	for {
		select {
		case <-time.After(a.interval):
			if n, err := a.plugin.Poll(a.authData); err != nil {
				log.Print("Error while polling ", a.authData.AccountId, ": ", err)
				continue
			} else if n != nil {
				postWatch <- &PostWatch{notifications: n, appId: a.plugin.ApplicationId()}
			}
		}
	}
}
