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
	watcher   *accounts.Watcher
	authData  accounts.AuthData
	plugin    plugins.Plugin
	interval  time.Duration
	postWatch chan *PostWatch
	authChan  chan accounts.AuthData
}

const DEFAULT_INTERVAL = time.Duration(60 * time.Second)

func NewAccountManager(watcher *accounts.Watcher, postWatch chan *PostWatch, plugin plugins.Plugin) *AccountManager {
	return &AccountManager{
		watcher:   watcher,
		plugin:    plugin,
		interval:  DEFAULT_INTERVAL,
		postWatch: postWatch,
		authChan:  make(chan accounts.AuthData, 1),
	}
}

func (a *AccountManager) Delete() {
	close(a.authChan)
}

func (a *AccountManager) Loop() {
	var ok bool
	if a.authData, ok = <- a.authChan; !ok {
		return
	}
L:
	for {
		log.Println("Polling set to", a.interval, "for", a.authData.AccountId)
		select {
		case <-time.After(a.interval):
			a.poll()
		case a.authData, ok = <-a.authChan:
			if !ok {
				break L
			}
			a.poll()
		}
	}
	log.Printf("Ending poll loop for account %d", a.authData.AccountId)
}

func (a *AccountManager) poll() {
	if !a.authData.Enabled {
		log.Println("Account", a.authData.AccountId, "no longer enabled")
		return
	}

	if a.authData.Error != nil {
		log.Println("Account", a.authData.AccountId, "failed to authenticate:", a.authData.Error)
		return
	}

	if n, err := a.plugin.Poll(&a.authData); err != nil {
		log.Print("Error while polling ", a.authData.AccountId, ": ", err)
		// penalizing the next poll
		a.interval += DEFAULT_INTERVAL

		// If the error indicates that the authentication
		// token has expired, request reauthentication and
		// mark data as disabled.
		if err == plugins.ErrTokenExpired {
			a.watcher.Refresh(a.authData.AccountId)
			a.authData.Enabled = false
		}
	} else if len(n) > 0 {
		// on success we reset the timeout to the default interval
		a.interval = DEFAULT_INTERVAL
		a.postWatch <- &PostWatch{messages: n, appId: a.plugin.ApplicationId()}
	}
}

func (a *AccountManager) updateAuthData(authData accounts.AuthData) {
	a.authChan <- authData
}
