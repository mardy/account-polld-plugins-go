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

	"launchpad.net/account-polld/plugins"
	"launchpad.net/account-polld/plugins/gmail"
)

type Account struct {
	id           string
	plugin       plugins.Plugin
	tokens       plugins.AuthTokens
	interval     time.Duration
	accountWatch chan plugins.AuthTokens
	installWatch chan bool
}

func NewAccount(id, plugin string) *Account {
	var p plugins.Plugin
	var interval time.Duration

	switch plugin {
	case "gmail":
		p = gmail.New()
		interval = 10 * time.Second
	}
	return &Account{
		id:           id,
		plugin:       p,
		interval:     interval,
		accountWatch: make(chan plugins.AuthTokens),
		installWatch: make(chan bool),
	}
}

func (a *Account) Delete() {
	close(a.accountWatch)
	close(a.installWatch)
}

func (a *Account) Loop(postWatch chan *PostWatch) {
	log.Println("Polling set to", a.interval)
	for {
		select {
		case t := <-a.accountWatch:
			a.tokens = t
		case <-time.After(a.interval):
			if n, err := a.plugin.Poll(a.tokens); err != nil {
				log.Print("Error while polling ", a.id, ": ", err)
				continue
			} else if n != nil {
				postWatch <- &PostWatch{notifications: n, appId: a.plugin.ApplicationId()}
			}
		}
	}
}
