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
	"os"
	"strconv"
	"time"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
	"launchpad.net/ubuntu-push/click"
)

type AccountManager struct {
	watcher   *accounts.Watcher
	authData  accounts.AuthData
	plugin    plugins.Plugin
	interval  time.Duration
	postWatch chan *PostWatch
	authChan  chan accounts.AuthData
}

var (
	pollInterval = time.Duration(5 * time.Minute)
	maxInterval  = time.Duration(20 * time.Minute)
)

func init() {
	if intervalEnv := os.Getenv("ACCOUNT_POLLD_POLL_INTERVAL_MINUTES"); intervalEnv != "" {
		if interval, err := strconv.ParseInt(intervalEnv, 0, 0); err == nil {
			pollInterval = time.Duration(interval) * time.Minute
		}
	}
}

func NewAccountManager(watcher *accounts.Watcher, postWatch chan *PostWatch, plugin plugins.Plugin) *AccountManager {
	return &AccountManager{
		watcher:   watcher,
		plugin:    plugin,
		interval:  pollInterval,
		postWatch: postWatch,
		authChan:  make(chan accounts.AuthData, 1),
	}
}

func (a *AccountManager) Delete() {
	close(a.authChan)
}

func (a *AccountManager) Loop() {
	var ok bool
	if a.authData, ok = <-a.authChan; !ok {
		return
	}
	// This is an initial out of loop poll
	a.poll()
L:
	for {
		log.Println("Next poll set to", a.interval, "for account", a.authData.AccountId)
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
	log.Println("Polling account", a.authData.AccountId)
	if false { // !isClickInstalled(a.plugin.ApplicationId()) {
		log.Println(
			"Skipping account", a.authData.AccountId, "as target click",
			a.plugin.ApplicationId(), "is not installed")
		return
	}

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
		if a.interval.Minutes() < maxInterval.Minutes() {
			a.interval += pollInterval
		}

		// If the error indicates that the authentication
		// token has expired, request reauthentication and
		// mark data as disabled.
		if err == plugins.ErrTokenExpired {
			a.watcher.Refresh(a.authData.AccountId)
			a.authData.Enabled = false
		}
	} else {
		log.Println("Account", a.authData.AccountId, "has", len(n), "updates to report")
		if len(n) > 0 {
			a.postWatch <- &PostWatch{messages: n, appId: a.plugin.ApplicationId()}
		}
		// on success we reset the timeout to the default interval
		a.interval = pollInterval
	}
}

func (a *AccountManager) updateAuthData(authData accounts.AuthData) {
	a.authChan <- authData
}

func isClickInstalled(appId plugins.ApplicationId) bool {
	user, err := click.User()
	if err != nil {
		log.Println("User instance for click cannot be created to determine if click application", appId, "was installed")
		return false
	}

	app, err := click.ParseAppId(string(appId))
	if err != nil {
		log.Println("Could not parse APP_ID for", appId)
		return false
	}

	return user.Installed(app, false)
}
