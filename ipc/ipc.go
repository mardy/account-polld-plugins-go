/*
 Copyright 2014 Canonical Ltd.

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

package ipc

import "C"
import (
	"encoding/json"
	"log"
	"os"
)

type Ipc struct {
	C       chan AuthData
	input *json.Decoder
	output *json.Encoder
}

type AuthData struct {
	ApplicationId string
	AccountId   uint
	ServiceName string
	ServiceType string
	Error       error
	Enabled     bool

	ClientId     string
	ClientSecret string
	AccessToken  string
	TokenSecret  string
	Secret       string
	UserName     string
}

type JsonInputMessage struct {
	ApplicationId string
	AccountId uint
	Auth map[string]interface{}
}

func NewIpc() *Ipc {
	w := new(Ipc)
	w.input = json.NewDecoder(os.Stdin)
	w.output = json.NewEncoder(os.Stdout)

	w.C = make(chan AuthData)

	return w
}

func (w *Ipc) Run() {
	for {
		var msg JsonInputMessage
		if err := w.input.Decode(&msg); err != nil {
			log.Println(err)
			return
		}
		log.Println("Got appId " + msg.ApplicationId)

		var data AuthData
		data.ApplicationId = msg.ApplicationId
		data.AccountId = msg.AccountId
		if v, ok := msg.Auth["accessToken"]; ok {
			data.AccessToken = v.(string)
		}
		if v, ok := msg.Auth["clientId"]; ok {
			data.ClientId = v.(string)
		}
		w.C <- data
	}
}

/*
func authCallback(watcher unsafe.Pointer, accountId C.uint, serviceType *C.char, serviceName *C.char, error *C.GError, enabled C.int, clientId, clientSecret, accessToken, tokenSecret *C.char, userName *C.char, secret *C.char, userData unsafe.Pointer) {
	// Ideally the first argument would be of type
	// *C.AccountIpc, but that fails with Go 1.2.
	authChannelsLock.Lock()
	ch := authChannels[(*C.AccountIpc)(watcher)]
	authChannelsLock.Unlock()
	if ch == nil {
		// Log the error
		return
	}

	var data AuthData
	data.AccountId = uint(accountId)
	data.ServiceName = C.GoString(serviceName)
	data.ServiceType = C.GoString(serviceType)
	if error != nil {
		data.Error = errors.New(C.GoString((*C.char)(error.message)))
	}
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
	if secret != nil {
		data.Secret = C.GoString(secret)
	}
	if userName != nil {
		data.UserName = C.GoString(userName)
	}
	ch <- data
}
*/
