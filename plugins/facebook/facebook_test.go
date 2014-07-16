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
package facebook

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

// closeWraper adds a dummy Close() method to a reader
type closeWrapper struct {
	io.Reader
}

func (r closeWrapper) Close() error {
	return nil
}

const (
	errorBody = `
{
  "error": {
    "message": "Message describing the error",
    "type": "OAuthException",
    "code": 190 ,
    "error_subcode": 460
  }
}`
	notificationsBody = `
{
  "data": [
    {
      "id": "notif_id",
      "from": {
        "id": "sender_id",
        "name": "Sender"
      },
      "to": {
        "id": "recipient_id",
        "name": "Recipient"
      },
      "created_time": "2014-07-12T09:51:57+0000",
      "updated_time": "2014-07-12T09:51:57+0000",
      "title": "Sender posted on your timeline: \"The message...\"",
      "link": "http://www.facebook.com/recipient/posts/id",
      "application": {
        "name": "Wall",
        "namespace": "wall",
        "id": "2719290516"
      },
      "unread": 1
    },
    {
      "id": "notif_1105650586_80600069",
      "from": {
        "id": "sender2_id",
        "name": "Sender2"
      },
      "to": {
        "id": "recipient_id",
        "name": "Recipient"
      },
      "created_time": "2014-07-08T06:17:52+0000",
      "updated_time": "2014-07-08T06:17:52+0000",
      "title": "Sender2's birthday was on July 7.",
      "link": "http://www.facebook.com/profile.php?id=xxx&ref=brem",
      "application": {
        "name": "Gifts",
        "namespace": "superkarma",
        "id": "329122197162272"
      },
      "unread": 1,
      "object": {
        "id": "sender2_id",
        "name": "Sender2"
      }
    }
  ],
  "paging": {
    "previous": "https://graph.facebook.com/v2.0/recipient/notifications?limit=5000&since=1405158717&__paging_token=enc_AewDzwIQmWOwPNO-36GaZsaJAog8l93HQ7uLEO-gp1Tb6KCiolXfzMCcGY2KjrJJsDJXdDmNJObICr5dewfMZgGs",
    "next": "https://graph.facebook.com/v2.0/recipient/notifications?limit=5000&until=1404705077&__paging_token=enc_Aewlhut5DQyhqtLNr7pLCMlYU012t4XY7FOt7cooz4wsWIWi-Jqz0a0IDnciJoeLu2vNNQkbtOpCmEmsVsN4hkM4"
  },
  "summary": [
  ]
}
`
)

func TestParseNotifications(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body: closeWrapper{bytes.NewReader([]byte(notificationsBody))},
	}
	p := &fbPlugin{}
	notifications, err := p.parseResponse(resp)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if len(notifications) != 2 {
		t.Fatal("Expected 2 notifications, got ", len(notifications))
	}
	if notifications[0].Card.Summary != "Sender posted on your timeline: \"The message...\"" {
		t.Error("Bad summary for first notification:", notifications[0].Card.Summary)
	}
	if notifications[1].Card.Summary != "Sender2's birthday was on July 7." {
		t.Error("Bad summary for second notification:", notifications[0].Card.Summary)
	}
	if p.lastUpdate != "2014-07-12T09:51:57+0000" {
		t.Error("Unexpected last update time:", p.lastUpdate)
	}
}

func TestIgnoreOldNotifications(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body: closeWrapper{bytes.NewReader([]byte(notificationsBody))},
	}
	p := &fbPlugin{lastUpdate: "2014-07-08T06:17:52+0000"}
	notifications, err := p.parseResponse(resp)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	if len(notifications) != 1 {
		t.Fatal("Expected 1 notification, got ", len(notifications))
	}
	if notifications[0].Card.Summary != "Sender posted on your timeline: \"The message...\"" {
		t.Error("Bad summary for first notification:", notifications[0].Card.Summary)
	}
	if p.lastUpdate != "2014-07-12T09:51:57+0000" {
		t.Error("Unexpected last update time:", p.lastUpdate)
	}
}

func TestErrorResponse(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body: closeWrapper{bytes.NewReader([]byte(errorBody))},
	}
	p := &fbPlugin{}
	notifications, err := p.parseResponse(resp)
	if err == nil {
		t.Fatal("Expected parseResponse to return an error.")
	}
	if notifications != nil {
		t.Error("Expected notifications to be nil on error.")
	}
	graphErr := err.(*GraphError)
	if graphErr.Message != "Message describing the error" {
		t.Errorf("Unexpected error message: '%s'", graphErr.Message)
	}
}
