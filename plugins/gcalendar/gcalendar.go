/*
 Copyright 2016 Canonical Ltd.

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
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"log"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
)

const (
	APP_ID      = "com.ubuntu.calendar_calendar"
	pluginName  = "gcalendar"
)

var baseUrl, _ = url.Parse("https://www.googleapis.com/calendar/v3/calendars/primary/events")


type GCalendarPlugin struct {
	accountId   uint
}

func lastSyncDateFromPersist(accountId uint) (lastSyncDate string, err error) {
	err = plugins.FromPersist(pluginName, accountId, &lastSyncDate)
	if err != nil {
		return "", err
	}
    return lastSyncDate, nil
}

func persist(accountId uint, lastSyncDate string) (err error) {
	err = plugins.Persist(pluginName, accountId, lastSyncDate)
	if err != nil {
		log.Print("calendar plugin ", accountId, ": failed to save state: ", err)
        return err
	}
	return nil
}

func New(accountId uint) *GCalendarPlugin {
	return &GCalendarPlugin{accountId: accountId}
}

func (p *GCalendarPlugin) ApplicationId() plugins.ApplicationId {
	return plugins.ApplicationId(APP_ID)
}

func (p *GCalendarPlugin) Poll(authData *accounts.AuthData) ([]*plugins.PushMessageBatch, error) {
	// This envvar check is to ease testing.
	if token := os.Getenv("ACCOUNT_POLLD_TOKEN_GCALENDAR"); token != "" {
		authData.AccessToken = token
	}

    lastSyncDate, err := lastSyncDateFromPersist(p.accountId)
	if err != nil {
		log.Print("calendar plugin ", p.accountId, ": cannot load previous sync date: ", err)
	} else {
		log.Print("calendar plugin ", p.accountId, ": last sync date: ", lastSyncDate)
	}

	resp, err := p.requestChanges(authData.AccessToken, lastSyncDate)
	if err != nil {
		return nil, err
	}

    messages, err := p.parseChangesResponse(resp)
	if err != nil {
		return nil, err
	}

    if len(messages) > 0 {
        // Update last sync date
        lastSyncDate = time.Now().UTC().Format(time.RFC3339)
        persist(p.accountId, lastSyncDate)
        log.Print("Request calendar sync")
    } else {
        log.Print("No sync is required.")
    }

    return nil, nil
}


func (p *GCalendarPlugin) parseChangesResponse(resp *http.Response) ([]event, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp errorResp
		if err := decoder.Decode(&errResp); err != nil {
			return nil, err
		}
		if errResp.Err.Code == 401 {
			return nil, plugins.ErrTokenExpired
		}
		return nil, nil
	}

	var events eventList
	if err := decoder.Decode(&events); err != nil {
		return nil, err
	}

    for _, ev := range events.Events {
        log.Print("Found event: ", ev.Etag, ev.Summary)
	}
	
	return events.Events, nil
}

func (p *GCalendarPlugin) requestChanges(accessToken string, lastSyncDate string) (*http.Response, error) {
	u, err := baseUrl.Parse("messages")
	if err != nil {
		return nil, err
	}

    //GET https://www.googleapis.com/calendar/v3/calendars/primary/events?showDeleted=true&singleEvents=true&updatedMin=2016-04-06T10%3A00%3A00.00Z&fields=description%2Citems(description%2Cetag%2Csummary)&key={YOUR_API_KEY}

	query := u.Query()
    query.Add("showDeleted", "true")
    query.Add("singleEvents", "true")
    query.Add("fields", "description,items(summary,etag)")
    query.Add("maxResults", "1")
    if len(lastSyncDate) > 0 {   
        query.Add("updatedMin", lastSyncDate)
    }
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	return http.DefaultClient.Do(req)
}
