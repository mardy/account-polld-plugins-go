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
package twitter

import (
	"bytes"
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
  "errors": [
    {
      "message":"Sorry, that page does not exist",
      "code":34
    }
  ]
}`
	tokenExpiredErrorBody = `
{
  "errors": [
    {
      "message":"Invalid or expired token",
      "code":89
    }
  ]
}`
	statusesBody = `
[
  {
    "coordinates": null,
    "favorited": false,
    "truncated": false,
    "created_at": "Mon Sep 03 13:24:14 +0000 2012",
    "id_str": "242613977966850048",
    "entities": {
      "urls": [

      ],
      "hashtags": [

      ],
      "user_mentions": [
        {
          "name": "Jason Costa",
          "id_str": "14927800",
          "id": 14927800,
          "indices": [
            0,
            11
          ],
          "screen_name": "jasoncosta"
        },
        {
          "name": "Matt Harris",
          "id_str": "777925",
          "id": 777925,
          "indices": [
            12,
            26
          ],
          "screen_name": "themattharris"
        },
        {
          "name": "ThinkWall",
          "id_str": "117426578",
          "id": 117426578,
          "indices": [
            109,
            119
          ],
          "screen_name": "thinkwall"
        }
      ]
    },
    "in_reply_to_user_id_str": "14927800",
    "contributors": null,
    "text": "@jasoncosta @themattharris Hey! Going to be in Frisco in October. Was hoping to have a meeting to talk about @thinkwall if you're around?",
    "retweet_count": 0,
    "in_reply_to_status_id_str": null,
    "id": 242613977966850048,
    "geo": null,
    "retweeted": false,
    "in_reply_to_user_id": 14927800,
    "place": null,
    "user": {
      "profile_sidebar_fill_color": "EEEEEE",
      "profile_sidebar_border_color": "000000",
      "profile_background_tile": false,
      "name": "Andrew Spode Miller",
      "profile_image_url": "http://a0.twimg.com/profile_images/1227466231/spode-balloon-medium_normal.jpg",
      "created_at": "Mon Sep 22 13:12:01 +0000 2008",
      "location": "London via Gravesend",
      "follow_request_sent": false,
      "profile_link_color": "F31B52",
      "is_translator": false,
      "id_str": "16402947",
      "entities": {
        "url": {
          "urls": [
            {
              "expanded_url": null,
              "url": "http://www.linkedin.com/in/spode",
              "indices": [
                0,
                32
              ]
            }
          ]
        },
        "description": {
          "urls": [

          ]
        }
      },
      "default_profile": false,
      "contributors_enabled": false,
      "favourites_count": 16,
      "url": "http://www.linkedin.com/in/spode",
      "profile_image_url_https": "https://si0.twimg.com/profile_images/1227466231/spode-balloon-medium_normal.jpg",
      "utc_offset": 0,
      "id": 16402947,
      "profile_use_background_image": false,
      "listed_count": 129,
      "profile_text_color": "262626",
      "lang": "en",
      "followers_count": 2013,
      "protected": false,
      "notifications": null,
      "profile_background_image_url_https": "https://si0.twimg.com/profile_background_images/16420220/twitter-background-final.png",
      "profile_background_color": "FFFFFF",
      "verified": false,
      "geo_enabled": true,
      "time_zone": "London",
      "description": "Co-Founder/Dev (PHP/jQuery) @justFDI. Run @thinkbikes and @thinkwall for events. Ex tech journo, helps run @uktjpr. Passion for Linux and customises everything.",
      "default_profile_image": false,
      "profile_background_image_url": "http://a0.twimg.com/profile_background_images/16420220/twitter-background-final.png",
      "statuses_count": 11550,
      "friends_count": 770,
      "following": null,
      "show_all_inline_media": true,
      "screen_name": "spode"
    },
    "in_reply_to_screen_name": "jasoncosta",
    "source": "<a href=\"http://www.journotwit.com\" rel=\"nofollow\">JournoTwit</a>",
    "in_reply_to_status_id": null
  },
  {
    "coordinates": {
      "coordinates": [
        121.0132101,
        14.5191613
      ],
      "type": "Point"
    },
    "favorited": false,
    "truncated": false,
    "created_at": "Mon Sep 03 08:08:02 +0000 2012",
    "id_str": "242534402280783873",
    "entities": {
      "urls": [

      ],
      "hashtags": [
        {
          "text": "twitter",
          "indices": [
            49,
            57
          ]
        }
      ],
      "user_mentions": [
        {
          "name": "Jason Costa",
          "id_str": "14927800",
          "id": 14927800,
          "indices": [
            14,
            25
          ],
          "screen_name": "jasoncosta"
        }
      ]
    },
    "in_reply_to_user_id_str": null,
    "contributors": null,
    "text": "Got the shirt @jasoncosta thanks man! Loving the #twitter bird on the shirt :-)",
    "retweet_count": 0,
    "in_reply_to_status_id_str": null,
    "id": 242534402280783873,
    "geo": {
      "coordinates": [
        14.5191613,
        121.0132101
      ],
      "type": "Point"
    },
    "retweeted": false,
    "in_reply_to_user_id": null,
    "place": null,
    "user": {
      "profile_sidebar_fill_color": "EFEFEF",
      "profile_sidebar_border_color": "EEEEEE",
      "profile_background_tile": true,
      "name": "Mikey",
      "profile_image_url": "http://a0.twimg.com/profile_images/1305509670/chatMikeTwitter_normal.png",
      "created_at": "Fri Jun 20 15:57:08 +0000 2008",
      "location": "Singapore",
      "follow_request_sent": false,
      "profile_link_color": "009999",
      "is_translator": false,
      "id_str": "15181205",
      "entities": {
        "url": {
          "urls": [
            {
              "expanded_url": null,
              "url": "http://about.me/michaelangelo",
              "indices": [
                0,
                29
              ]
            }
          ]
        },
        "description": {
          "urls": [

          ]
        }
      },
      "default_profile": false,
      "contributors_enabled": false,
      "favourites_count": 11,
      "url": "http://about.me/michaelangelo",
      "profile_image_url_https": "https://si0.twimg.com/profile_images/1305509670/chatMikeTwitter_normal.png",
      "utc_offset": 28800,
      "id": 15181205,
      "profile_use_background_image": true,
      "listed_count": 61,
      "profile_text_color": "333333",
      "lang": "en",
      "followers_count": 577,
      "protected": false,
      "notifications": null,
      "profile_background_image_url_https": "https://si0.twimg.com/images/themes/theme14/bg.gif",
      "profile_background_color": "131516",
      "verified": false,
      "geo_enabled": true,
      "time_zone": "Hong Kong",
      "description": "Android Applications Developer,  Studying Martial Arts, Plays MTG, Food and movie junkie",
      "default_profile_image": false,
      "profile_background_image_url": "http://a0.twimg.com/images/themes/theme14/bg.gif",
      "statuses_count": 11327,
      "friends_count": 138,
      "following": null,
      "show_all_inline_media": true,
      "screen_name": "mikedroid"
    },
    "in_reply_to_screen_name": null,
    "source": "<a href=\"http://twitter.com/download/android\" rel=\"nofollow\">Twitter for Android</a>",
    "in_reply_to_status_id": null
  }
]`
	directMessagesBody = `
[
{
    "created_at": "Mon Aug 27 17:21:03 +0000 2012",
    "entities": {
        "hashtags": [],
        "urls": [],
        "user_mentions": []
    },
    "id": 240136858829479936,
    "id_str": "240136858829479936",
    "recipient": {
        "contributors_enabled": false,
        "created_at": "Thu Aug 23 19:45:07 +0000 2012",
        "default_profile": false,
        "default_profile_image": false,
        "description": "Keep calm and test",
        "favourites_count": 0,
        "follow_request_sent": false,
        "followers_count": 0,
        "following": false,
        "friends_count": 10,
        "geo_enabled": true,
        "id": 776627022,
        "id_str": "776627022",
        "is_translator": false,
        "lang": "en",
        "listed_count": 0,
        "location": "San Francisco, CA",
        "name": "Mick Jagger",
        "notifications": false,
        "profile_background_color": "000000",
        "profile_background_image_url": "http://a0.twimg.com/profile_background_images/644522235/cdjlccey99gy36j3em67.jpeg",
        "profile_background_image_url_https": "https://si0.twimg.com/profile_background_images/644522235/cdjlccey99gy36j3em67.jpeg",
        "profile_background_tile": true,
        "profile_image_url": "http://a0.twimg.com/profile_images/2550226257/y0ef5abcx5yrba8du0sk_normal.jpeg",
        "profile_image_url_https": "https://si0.twimg.com/profile_images/2550226257/y0ef5abcx5yrba8du0sk_normal.jpeg",
        "profile_link_color": "000000",
        "profile_sidebar_border_color": "000000",
        "profile_sidebar_fill_color": "000000",
        "profile_text_color": "000000",
        "profile_use_background_image": false,
        "protected": false,
        "screen_name": "s0c1alm3dia",
        "show_all_inline_media": false,
        "statuses_count": 0,
        "time_zone": "Pacific Time (US & Canada)",
        "url": "http://cnn.com",
        "utc_offset": -28800,
        "verified": false
    },
    "recipient_id": 776627022,
    "recipient_screen_name": "s0c1alm3dia",
    "sender": {
        "contributors_enabled": true,
        "created_at": "Sat May 09 17:58:22 +0000 2009",
        "default_profile": false,
        "default_profile_image": false,
        "description": "I taught your phone that thing you like.  The Mobile Partner Engineer @Twitter. ",
        "favourites_count": 584,
        "follow_request_sent": false,
        "followers_count": 10621,
        "following": false,
        "friends_count": 1181,
        "geo_enabled": true,
        "id": 38895958,
        "id_str": "38895958",
        "is_translator": false,
        "lang": "en",
        "listed_count": 190,
        "location": "San Francisco",
        "name": "Sean Cook",
        "notifications": false,
        "profile_background_color": "1A1B1F",
        "profile_background_image_url": "http://a0.twimg.com/profile_background_images/495742332/purty_wood.png",
        "profile_background_image_url_https": "https://si0.twimg.com/profile_background_images/495742332/purty_wood.png",
        "profile_background_tile": true,
        "profile_image_url": "http://a0.twimg.com/profile_images/1751506047/dead_sexy_normal.JPG",
        "profile_image_url_https": "https://si0.twimg.com/profile_images/1751506047/dead_sexy_normal.JPG",
        "profile_link_color": "2FC2EF",
        "profile_sidebar_border_color": "181A1E",
        "profile_sidebar_fill_color": "252429",
        "profile_text_color": "666666",
        "profile_use_background_image": true,
        "protected": false,
        "screen_name": "theSeanCook",
        "show_all_inline_media": true,
        "statuses_count": 2608,
        "time_zone": "Pacific Time (US & Canada)",
        "url": null,
        "utc_offset": -28800,
        "verified": false
    },
    "sender_id": 38895958,
    "sender_screen_name": "theSeanCook",
    "text": "booyakasha"
}
]
`
)

func (s S) TestParseStatuses(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       closeWrapper{bytes.NewReader([]byte(statusesBody))},
	}
	p := &twitterPlugin{}
	batch, err := p.parseStatuses(resp)
	c.Assert(err, IsNil)
	c.Assert(batch, NotNil)
	messages := batch.Messages
	c.Assert(len(messages), Equals, 2)
	c.Check(messages[0].Notification.Card.Summary, Equals, "Andrew Spode Miller. @spode")
	c.Check(messages[0].Notification.Card.Body, Equals, "@jasoncosta @themattharris Hey! Going to be in Frisco in October. Was hoping to have a meeting to talk about @thinkwall if you're around?")
	c.Check(messages[0].Notification.Card.Icon, Equals, "http://a0.twimg.com/profile_images/1227466231/spode-balloon-medium_normal.jpg")
	c.Assert(len(messages[0].Notification.Card.Actions), Equals, 1)
	c.Check(messages[0].Notification.Card.Actions[0], Equals, "https://mobile.twitter.com/spode/statuses/242613977966850048")
	c.Check(messages[1].Notification.Card.Summary, Equals, "Mikey. @mikedroid")
	c.Check(messages[1].Notification.Card.Body, Equals, "Got the shirt @jasoncosta thanks man! Loving the #twitter bird on the shirt :-)")
	c.Check(messages[1].Notification.Card.Icon, Equals, "http://a0.twimg.com/profile_images/1305509670/chatMikeTwitter_normal.png")
	c.Assert(len(messages[1].Notification.Card.Actions), Equals, 1)
	c.Check(messages[1].Notification.Card.Actions[0], Equals, "https://mobile.twitter.com/mikedroid/statuses/242534402280783873")
	c.Check(p.lastMentionId, Equals, int64(242613977966850048))
}

func (s S) TestParseStatusesError(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       closeWrapper{bytes.NewReader([]byte(errorBody))},
	}
	p := &twitterPlugin{}
	messages, err := p.parseStatuses(resp)
	c.Check(messages, IsNil)
	c.Assert(err, Not(IsNil))
	twErr := err.(*TwitterError)
	c.Assert(len(twErr.Errors), Equals, 1)
	c.Check(twErr.Errors[0].Message, Equals, "Sorry, that page does not exist")
	c.Check(twErr.Errors[0].Code, Equals, 34)
}

func (s S) TestParseStatusesTokenExpiredError(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       closeWrapper{bytes.NewReader([]byte(tokenExpiredErrorBody))},
	}
	p := &twitterPlugin{}
	messages, err := p.parseStatuses(resp)
	c.Check(messages, IsNil)
	c.Assert(err, Equals, plugins.ErrTokenExpired)
}

func (s S) TestParseDirectMessages(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       closeWrapper{bytes.NewReader([]byte(directMessagesBody))},
	}
	p := &twitterPlugin{}
	batch, err := p.parseDirectMessages(resp)
	c.Assert(err, IsNil)
	c.Assert(batch, NotNil)
	messages := batch.Messages
	c.Assert(len(messages), Equals, 1)
	c.Check(messages[0].Notification.Card.Summary, Equals, "Sean Cook. @theSeanCook")
	c.Check(messages[0].Notification.Card.Body, Equals, "booyakasha")
	c.Check(messages[0].Notification.Card.Icon, Equals, "http://a0.twimg.com/profile_images/1751506047/dead_sexy_normal.JPG")
	c.Assert(len(messages[0].Notification.Card.Actions), Equals, 1)
	c.Check(messages[0].Notification.Card.Actions[0], Equals, "https://mobile.twitter.com/theSeanCook/messages")
	c.Check(p.lastDirectMessageId, Equals, int64(240136858829479936))
}

func (s S) TestParseDirectMessagesError(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       closeWrapper{bytes.NewReader([]byte(errorBody))},
	}
	p := &twitterPlugin{}
	messages, err := p.parseDirectMessages(resp)
	c.Check(messages, IsNil)
	c.Assert(err, Not(IsNil))
	twErr := err.(*TwitterError)
	c.Assert(len(twErr.Errors), Equals, 1)
	c.Check(twErr.Errors[0].Message, Equals, "Sorry, that page does not exist")
	c.Check(twErr.Errors[0].Code, Equals, 34)
}

func (s S) TestParseDirectMessagesTokenExpiredError(c *C) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       closeWrapper{bytes.NewReader([]byte(tokenExpiredErrorBody))},
	}
	p := &twitterPlugin{}
	messages, err := p.parseDirectMessages(resp)
	c.Check(messages, IsNil)
	c.Assert(err, Equals, plugins.ErrTokenExpired)
}
