package qtcontact

// #cgo CXXFLAGS: -std=c++0x -pedantic-errors -Wall -fno-strict-aliasing -I/usr/share/c++/4.8
// #cgo LDFLAGS: -lstdc++
// #cgo pkg-config: Qt5Core Qt5Contacts
// #include "qtcontacts.h"
import "C"

import (
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
	case <-time.After(10 * time.Second):
		return ""
	case path := <-avatarPathChan:
		return path
	}
}
