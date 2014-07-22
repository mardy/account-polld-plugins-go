/*
 Copyright 2014 Canonical Ltd.
 Authors: James Henstridge <james.henstridge@canonical.com>

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

package accounts

/*
#cgo pkg-config: glib-2.0 libaccounts-glib libsignon-glib
#include <stdlib.h>
#include <glib.h>
#include "account-watcher.h"

AccountWatcher *watch_for_services(void *array_of_service_names, int length);
*/
import "C"
import (
	"sync"
	"unsafe"
)

type Watcher struct {
	C <-chan AuthData
	w *C.AccountWatcher
}

type AuthData struct {
	AccountId   uint
	ServiceName string
	Enabled     bool

	ClientId     string
	ClientSecret string
	AccessToken  string
	TokenSecret  string
}

var (
	mainLoopOnce sync.Once
	authChannels = make(map[*C.AccountWatcher]chan<- AuthData)
	authChannelsLock sync.Mutex
)

func startMainLoop() {
	mainLoopOnce.Do(func() {
		mainLoop := C.g_main_loop_new(nil, C.gboolean(1))
		go C.g_main_loop_run(mainLoop)
	})
}

// NewWatcher creates a new account watcher for the given service names
func NewWatcher(serviceNames... string) *Watcher {
	watcher := new(Watcher)
	watcher.w = C.watch_for_services(unsafe.Pointer(&serviceNames[0]), C.int(len(serviceNames)))

	ch := make(chan AuthData)
	watcher.C = ch
	authChannelsLock.Lock()
	authChannels[watcher.w] = ch
	authChannelsLock.Unlock()

	startMainLoop()

	return watcher
}

//export authCallback
func authCallback(watcher unsafe.Pointer, accountId C.uint, serviceName *C.char, enabled C.int, clientId, clientSecret, accessToken, tokenSecret *C.char, userData unsafe.Pointer) {
	// Ideally the first argument would be of type
	// *C.AccountWatcher, but that fails with Go 1.2.
	authChannelsLock.Lock()
	ch := authChannels[(*C.AccountWatcher)(watcher)]
	authChannelsLock.Unlock()
	if ch == nil {
		// Log the error
		return
	}

	var data AuthData
	data.AccountId = uint(accountId)
	data.ServiceName = C.GoString(serviceName)
	if enabled != 0 {
		data.Enabled = true
	}
	if clientId != nil {
		data.ClientId = C.GoString(clientId)
	}
	if clientSecret != nil {
		data.ClientSecret = C.GoString(clientSecret)
	}
	if accessToken != nil {
		data.AccessToken = C.GoString(accessToken)
	}
	if tokenSecret != nil {
		data.TokenSecret = C.GoString(tokenSecret)
	}
	ch <- data
}
