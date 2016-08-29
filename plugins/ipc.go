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

package plugins

import "C"
import (
	"encoding/json"
	"log"
	"os"
)

type Ipc struct {
	C      chan AuthData
	input  *json.Decoder
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

func NewIpc(authData chan AuthData) *Ipc {
	w := new(Ipc)
	w.input = json.NewDecoder(os.Stdin)
	w.output = json.NewEncoder(os.Stdout)

	w.C = authData

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
		if v, ok := msg.Auth["clientId"]; ok {
			data.ClientId = v.(string)
		}
		if v, ok := msg.Auth["clientSecret"]; ok {
			data.ClientSecret = v.(string)
		}
		if v, ok := msg.Auth["accessToken"]; ok {
			data.AccessToken = v.(string)
		}
		if v, ok := msg.Auth["tokenSecret"]; ok {
			data.TokenSecret = v.(string)
		}
		if v, ok := msg.Auth["secret"]; ok {
			data.Secret = v.(string)
		}
		if v, ok := msg.Auth["userName"]; ok {
			data.UserName = v.(string)
		}
		w.C <- data
	}
}

func (w *Ipc) PostMessages(batches []*PushMessageBatch) {
	var notifications []*PushMessage

	for _, batch := range batches {
		notifs := batch.Messages
		overflowing := len(notifs) > batch.Limit

		for i, n := range notifs {
			// Play sound and vibrate on first notif only.
			if i > 0 {
				n.Notification.Vibrate = false
				n.Notification.Sound = ""
			}

			// We're overflowing, so no popups.
			// See LP: #1527171
			if overflowing {
				n.Notification.Card.Popup = false
			}
		}

		if overflowing {
			n := batch.OverflowHandler(notifs)
			n.Notification.Card.Persist = false
			n.Notification.Vibrate = false
			notifs = append(notifs, n)
		}

		notifications = append(notifications, notifs...)
	}

	reply := make(map[string]interface{})
	reply["notifications"] = notifications
	w.output.Encode(reply)
}

func (w *Ipc) PostError(err error) {
	errorMap := make(map[string]string)
	errorMap["message"] = err.Error()
	if err == ErrTokenExpired {
		errorMap["code"] = "ERR_INVALID_AUTH"
	}
	reply := make(map[string]interface{})
	reply["error"] = errorMap
	w.output.Encode(reply)
}
