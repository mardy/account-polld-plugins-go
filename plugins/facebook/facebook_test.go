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
	"fmt"
	"io"
	"net/http"
	"testing"

	. "launchpad.net/gocheck"

	"launchpad.net/account-polld/plugins"
)

type S struct{}

func init() {
	Suite(S{})
}

func TestAll(t *testing.T) {
	TestingT(t)
}

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
    "message": "Unknown path components: /xyz",
    "type": "OAuthException",
    "code": 2500
  }
}`
	tokenExpiredErrorBody = `
{
  "error": {
    "message": "Error validating access token: Session has expired",
    "type": "OAuthException",
    "code": 190 ,
    "error_subcode": 463
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
	largeNotificationsBody = `
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
    },
    {
      "id": "notif_id_3",
      "from": {
        "id": "sender3_id",
        "name": "Sender3"
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
      "id": "notif_id_4",
      "from": {
        "id": "sender4_id",
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
      "unread": 1
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

	inboxBody = `
{
  "data": [
    {
      "unread": 1, 
      "unseen": 1, 
      "id": "445809168892281", 
      "updated_time": "2014-08-25T18:39:32+0000", 
      "comments": {
        "data": [
          {
            "id": "445809168892281_1408991972", 
            "from": {
              "id": "346217352202239", 
              "name": "Pollod Magnifico"
            }, 
            "message": "Hola mundo!", 
            "created_time": "2014-08-25T18:39:32+0000"
          }
        ], 
        "paging": {
          "previous": "https://graph.facebook.com/v2.0/445809168892281/comments?limit=1&since=1408991972&__paging_token=enc_Aew2kKJXEXzdm9k89DvLYz_y8nYxUbvElWcn6h_pKMRsoAPTPpkU7-AsGhkcYF6M1qbomOnFJf9ckL5J3hTltLFq", 
          "next": "https://graph.facebook.com/v2.0/445809168892281/comments?limit=1&until=1408991972&__paging_token=enc_Aewlixpk4h4Vq79-W1ixrTM6ONbsMUDrcj0vLABs34tbhWarfpQLf818uoASWNDEpQO4XEXh5HbgHpcCqnuNVEOR"
        }
      }
    },
    {
      "unread": 2, 
      "unseen": 1, 
      "id": "445809168892282", 
      "updated_time": "2014-08-25T18:39:32+0000", 
      "comments": {
        "data": [
          {
            "id": "445809168892282_1408991973", 
            "from": {
              "id": "346217352202239", 
              "name": "Pollitod Magnifico"
            }, 
            "message": "Hola!", 
            "created_time": "2014-08-25T18:39:32+0000"
          }
        ], 
        "paging": {
          "previous": "https://graph.facebook.com/v2.0/445809168892281/comments?limit=1&since=1408991972&__paging_token=enc_Aew2kKJXEXzdm9k89DvLYz_y8nYxUbvElWcn6h_pKMRsoAPTPpkU7-AsGhkcYF6M1qbomOnFJf9ckL5J3hTltLFq", 
          "next": "https://graph.facebook.com/v2.0/445809168892281/comments?limit=1&until=1408991972&__paging_token=enc_Aewlixpk4h4Vq79-W1ixrTM6ONbsMUDrcj0vLABs34tbhWarfpQLf818uoASWNDEpQO4XEXh5HbgHpcCqnuNVEOR"
        }
      }
    },
    {
      "unread": 2, 
      "unseen": 1, 
      "id": "445809168892283", 
      "updated_time": "2014-08-25T18:39:32+0000", 
      "comments": {
        "data": [
          {
            "id": "445809168892282_1408991973", 
            "from": {
              "id": "346217352202240", 
              "name": "A Friend"
            }, 
            "message": "mellon", 
            "created_time": "2014-08-25T18:39:32+0000"
          }
        ], 
        "paging": {
          "previous": "https://graph.facebook.com/v2.0/445809168892281/comments?limit=1&since=1408991972&__paging_token=enc_Aew2kKJXEXzdm9k89DvLYz_y8nYxUbvElWcn6h_pKMRsoAPTPpkU7-AsGhkcYF6M1qbomOnFJf9ckL5J3hTltLFq", 
          "next": "https://graph.facebook.com/v2.0/445809168892281/comments?limit=1&until=1408991972&__paging_token=enc_Aewlixpk4h4Vq79-W1ixrTM6ONbsMUDrcj0vLABs34tbhWarfpQLf818uoASWNDEpQO4XEXh5HbgHpcCqnuNVEOR"
        }
      }
    }


  ], 
  "paging": {
    "previous": "https://graph.facebook.com/v2.0/270128826512416/inbox?fields=unread,unseen,comments.limit(1)&limit=25&since=1408991972&__paging_token=enc_Aey99ACSOyZqN_7I-yWLnY8K3dqu4wVsx-Th3kMHMTMQ5VPbQRPgCQiJps0II1QAXDAVzHplqPS8yNgq8Zs_G2aK", 
    "next": "https://graph.facebook.com/v2.0/270128826512416/inbox?fields=unread,unseen,comments.limit(1)&limit=25&until=1408991972&__paging_token=enc_AewjHkk10NNjRCXJCoaP5hyf22kw-htwxsDaVOiLY-IiXxB99sKNGlfFFmkcG-VeMGUETI2agZGR_1IWP5W4vyPL"
  }, 
  "summary": {
    "unseen_count": 0, 
    "unread_count": 1, 
    "updated_time": "2014-08-25T19:05:49+0000"
  }
}
`
)

func (s S) TestParseNotifications(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       closeWrapper{bytes.NewReader([]byte(notificationsBody))},
	}
	p := &fbPlugin{}
	messages, err := p.parseResponse(resp)
	c.Assert(err, IsNil)
	c.Assert(len(messages), Equals, 2)
	fmt.Println(messages)
	c.Check(messages[0].Notification.Card.Summary, Equals, "Sender")
	c.Check(messages[0].Notification.Card.Body, Equals, "Sender posted on your timeline: \"The message...\"")
	c.Check(messages[1].Notification.Card.Summary, Equals, "Sender2")
	c.Check(messages[1].Notification.Card.Body, Equals, "Sender2's birthday was on July 7.")
	c.Check(p.state.lastUpdate, Equals, timeStamp("2014-07-12T09:51:57+0000"))
}

func (s S) TestParseLotsOfNotifications(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       closeWrapper{bytes.NewReader([]byte(largeNotificationsBody))},
	}
	p := &fbPlugin{}
	messages, err := p.parseResponse(resp)
	c.Assert(err, IsNil)
	c.Assert(len(messages), Equals, 3)
	fmt.Println(messages)
	c.Check(messages[0].Notification.Card.Summary, Equals, "Sender")
	c.Check(messages[0].Notification.Card.Body, Equals, "Sender posted on your timeline: \"The message...\"")
	c.Check(messages[1].Notification.Card.Summary, Equals, "Sender2")
	c.Check(messages[1].Notification.Card.Body, Equals, "Sender2's birthday was on July 7.")
	c.Check(messages[2].Notification.Card.Summary, Equals, "Multiple more notifications")
	c.Check(messages[2].Notification.Card.Body, Equals, "From Sender3, Sender2")
	c.Check(p.state.lastUpdate, Equals, timeStamp("2014-07-12T09:51:57+0000"))
}

func (s S) TestIgnoreOldNotifications(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       closeWrapper{bytes.NewReader([]byte(notificationsBody))},
	}
	p := &fbPlugin{state: fbState{lastUpdate: "2014-07-08T06:17:52+0000"}}
	messages, err := p.parseResponse(resp)
	c.Assert(err, IsNil)
	c.Assert(len(messages), Equals, 1)
	c.Check(messages[0].Notification.Card.Summary, Equals, "Sender")
	c.Check(messages[0].Notification.Card.Body, Equals, "Sender posted on your timeline: \"The message...\"")
	c.Check(p.state.lastUpdate, Equals, timeStamp("2014-07-12T09:51:57+0000"))
}

func (s S) TestErrorResponse(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       closeWrapper{bytes.NewReader([]byte(errorBody))},
	}
	p := &fbPlugin{}
	notifications, err := p.parseResponse(resp)
	c.Check(notifications, IsNil)
	c.Assert(err, Not(IsNil))
	graphErr := err.(*GraphError)
	c.Check(graphErr.Message, Equals, "Unknown path components: /xyz")
	c.Check(graphErr.Code, Equals, 2500)
}

func (s S) TestTokenExpiredErrorResponse(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       closeWrapper{bytes.NewReader([]byte(tokenExpiredErrorBody))},
	}
	p := &fbPlugin{}
	notifications, err := p.parseResponse(resp)
	c.Check(notifications, IsNil)
	c.Assert(err, Equals, plugins.ErrTokenExpired)
}

func (s S) TestParseInbox(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       closeWrapper{bytes.NewReader([]byte(inboxBody))},
	}
	p := &fbPlugin{}
	messages, err := p.parseInboxResponse(resp)
	c.Assert(err, IsNil)
	c.Assert(len(messages), Equals, 3)
	c.Check(messages[0].Notification.Card.Summary, Equals, "Pollod Magnifico")
	c.Check(messages[0].Notification.Card.Body, Equals, "Hola mundo!")
	c.Check(messages[1].Notification.Card.Summary, Equals, "Pollitod Magnifico")
	c.Check(messages[1].Notification.Card.Body, Equals, "Hola!")
	c.Check(messages[2].Notification.Card.Summary, Equals, "Multiple more messages")
	c.Check(messages[2].Notification.Card.Body, Equals, "From Pollitod Magnifico, A Friend")
	c.Check(p.state.lastInboxUpdate, Equals, timeStamp("2014-08-25T18:39:32+0000"))
}
