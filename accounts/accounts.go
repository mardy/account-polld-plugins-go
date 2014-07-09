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

typedef struct _AuthContext AuthContext;

AuthContext *watch_for_service(const char *service_name);
*/
import "C"
import "unsafe"

type AuthData struct {
	ClientId     string
	ClientSecret string
	AccessToken  string
}

var (
	mainLoop     *C.GMainLoop
	authChannels = make(map[*C.AuthContext]chan<- AuthData)
)

func WatchForService(serviceName string) <-chan AuthData {
	if mainLoop == nil {
		mainLoop = C.g_main_loop_new(nil, C.gboolean(1))
		go C.g_main_loop_run(mainLoop)
	}

	cService := C.CString(serviceName)
	defer C.free(unsafe.Pointer(cService))
	ctx := C.watch_for_service(cService)

	ch := make(chan AuthData)
	authChannels[ctx] = ch
	return ch
}

//export authLogin
func authLogin(user_data unsafe.Pointer, clientId *C.char, clientSecret *C.char, accessToken *C.char) {
	ctx := (*C.AuthContext)(user_data)
	ch := authChannels[ctx]
	if ch == nil {
		// Log the error
		return
	}

	var data AuthData
	if clientId != nil {
		data.ClientId = C.GoString(clientId)
	}
	if clientSecret != nil {
		data.ClientSecret = C.GoString(clientSecret)
	}
	if accessToken != nil {
		data.AccessToken = C.GoString(accessToken)
	}
	ch <- data
}
