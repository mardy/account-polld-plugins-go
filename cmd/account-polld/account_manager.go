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
	"sync"
	"time"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
)

type AccountManager struct {
	authData  accounts.AuthData
	authMutex *sync.Mutex
	plugin    plugins.Plugin
	interval  time.Duration
	postWatch chan *PostWatch
	terminate chan bool
}

const DEFAULT_INTERVAL = time.Duration(60 * time.Second)

func NewAccountManager(authData accounts.AuthData, postWatch chan *PostWatch, plugin plugins.Plugin) *AccountManager {
	return &AccountManager{
		plugin:    plugin,
		authData:  authData,
		authMutex: &sync.Mutex{},
		postWatch: postWatch,
		interval:  DEFAULT_INTERVAL,
		terminate: make(chan bool),
	}
}

func (a *AccountManager) Delete() {
	a.terminate <- true
}

func (a *AccountManager) Loop() {
	defer close(a.terminate)
L:
	for {
		log.Println("Polling set to", a.interval, "for", a.authData.AccountId)
		select {
		case <-time.After(a.interval):
			a.poll()
		case <-a.terminate:
			break L
		}
	}
	log.Printf("Ending poll loop for account %d", a.authData.AccountId)
}

func (a *AccountManager) poll() {
	a.authMutex.Lock()
	defer a.authMutex.Unlock()

	if !a.authData.Enabled {
		log.Println("Account", a.authData.AccountId, "no longer enabled")
		return
	}

	if n, err := a.plugin.Poll(&a.authData); err != nil {
		log.Print("Error while polling ", a.authData.AccountId, ": ", err)
		// penalizing the next poll
		a.interval += DEFAULT_INTERVAL
	} else if len(n) > 0 {
		// on success we reset the timeout to the default interval
		a.interval = DEFAULT_INTERVAL
		a.postWatch <- &PostWatch{messages: n, appId: a.plugin.ApplicationId()}
	}
}

func (a *AccountManager) updateAuthData(authData accounts.AuthData) {
	a.authMutex.Lock()
	defer a.authMutex.Unlock()
	a.authData = authData
}
