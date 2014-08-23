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

package qtcontact

// #cgo CXXFLAGS: -std=c++0x -pedantic-errors -Wall -fno-strict-aliasing -I/usr/share/c++/4.8
// #cgo LDFLAGS: -lstdc++
// #cgo pkg-config: Qt5Core Qt5Contacts
// #include "qtcontacts.h"
import "C"

import (
	"log"
	"sync"
	"time"
)

var (
	avatarPathChan chan string
	m              sync.Mutex
)

//export callback
func callback(path *C.char) {
	avatarPathChan <- C.GoString(path)
}

func MainLoopStart() {
	go C.mainloopStart()
}

// GetAvatar retrieves an avatar path for the specified email
// address. Multiple calls to this func will be in sync
func GetAvatar(emailAddress string) string {
	if emailAddress == "" {
		return ""
	}

	m.Lock()
	defer m.Unlock()

	avatarPathChan = make(chan string, 1)
	defer close(avatarPathChan)

	C.getAvatar(C.CString(emailAddress))

	select {
	case <-time.After(3 * time.Second):
		log.Println("Timeout while seeking avatar for", emailAddress)
		return ""
	case path := <-avatarPathChan:
		return path
	}
}
