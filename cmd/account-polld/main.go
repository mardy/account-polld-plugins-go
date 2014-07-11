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

	"launchpad.net/account-polld/plugins"
	"launchpad.net/go-dbus/v1"
)

type PostWatch struct {
	appId         plugins.ApplicationId
	notifications *[]plugins.Notification
}

func main() {
	// TODO NewAccount called here is just for playing purposes.
	a := NewAccount("sergiusens@gmail.com", "gmail")
	defer a.Delete()
	postWatch := make(chan *PostWatch)

	if bus, err := dbus.Connect(dbus.SessionBus); err != nil {
		log.Fatal("Cannot connect to bus", err)
	} else {
		go postOffice(bus, postWatch)
	}

	go a.Loop(postWatch)

	done := make(chan bool)
	<-done
}

func postOffice(bus *dbus.Connection, postWatch chan *PostWatch) {
	for post := range postWatch {
		for _, n := range *post.notifications {
			fmt.Println("Should be dispathing", n, "to the post office using", bus.UniqueName, "for", post.appId)
		}
	}
}
