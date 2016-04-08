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

package gcalendar

import (
	"fmt"
)


// eventList holds a response to call to Calendar.events: list
// defined in https://developers.google.com/google-apps/calendar/v3/reference/events/list#response
type eventList struct {
	// Messages holds a list of message.
	Events []event `json:"items"`
}

// event holds the event data response for a Calendar.event.
// The full definition of a message is defined in
// https://developers.google.com/google-apps/calendar/v3/reference/events#resource-representations
type event struct {
	// Id is the immutable ID of the message.
	Etag string `json:"etag"`
	// ThreadId is the ID of the thread the message belongs to.
	Summary string `json:"summary"`
}

func (e event) String() string {
	return fmt.Sprintf("Id: %s, snippet: '%s'\n", e.Etag, e.Summary)
}


type errorResp struct {
	Err struct {
		Code    uint64 `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Domain  string `json:"domain"`
			Reason  string `json:"reason"`
			Message string `json:"message"`
		} `json:"errors"`
	} `json:"error"`
}
