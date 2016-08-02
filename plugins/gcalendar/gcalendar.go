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
	"log"
	"net/http"
	"net/url"
	"os"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
	"launchpad.net/account-polld/syncmonitor"
)

const (
	APP_ID     = "com.ubuntu.calendar_calendar"
	pluginName = "gcalendar"
)

var baseUrl, _ = url.Parse("https://www.googleapis.com/calendar/v3/calendars/")

type GCalendarPlugin struct {
	accountId uint
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
		log.Print("calendar: Using token from: ACCOUNT_POLLD_TOKEN_GCALENDAR env var")
		authData.AccessToken = token
	}

	log.Print("calendar: Check calendar changes for account:", p.accountId)

	syncMonitor := syncmonitor.NewSyncMonitor()
	if syncMonitor == nil {
		log.Print("calendar: Sync monitor not available yet.")
		return nil, nil
	}

	state, err := syncMonitor.State()
	if err != nil {
		log.Print("calendar: Fail to retrieve sync monitor state ", err)
		return nil, nil
	}
	if state != "idle" {
		log.Print("calendar: Sync monitor is not on 'idle' state, try later!")
		return nil, nil
	}

	calendars, err := syncMonitor.ListCalendarsByAccount(p.accountId)
	if err != nil {
		log.Print("calendar: Calendar plugin ", p.accountId, ": cannot load calendars: ", err)
		return nil, nil
	}

	var calendarsToSync []string
	log.Print("calendar: Number of calendars for account:", p.accountId, " size:", len(calendars))

	for id, calendar := range calendars {
		lastSyncDate, err := syncMonitor.LastSyncDate(p.accountId, id)
		if err != nil {
			log.Print("\tcalendar: ", calendar, ", cannot load previous sync date: ", err, ". Try next time.")
			continue
		} else {
			log.Print("\tcalendar: ", calendar, " Id: ", id, ": last sync date: ", lastSyncDate)
		}

		var needSync bool
		needSync = (len(lastSyncDate) == 0)

		if !needSync {
			resp, err := p.requestChanges(authData.AccessToken, id, lastSyncDate)
			if err != nil {
				log.Print("\tcalendar: ERROR: Fail to query for changes: ", err)
				continue
			}

			messages, err := p.parseChangesResponse(resp)
			if err != nil {
				log.Print("\tcalendar: ERROR: Fail to parse changes: ", err)
				if err == plugins.ErrTokenExpired {
					log.Print("\t\tcalendar: Abort poll")
					return nil, err
				} else {
					continue
				}
			}
			needSync = (len(messages) > 0)
		}

		if needSync {
			log.Print("\tcalendar: Calendar needs sync: ", calendar)
			calendarsToSync = append(calendarsToSync, id)
		} else {
			log.Print("\tcalendar: Found no calendar updates for account: ", p.accountId, " calendar: ", calendar)
		}
	}

	if len(calendarsToSync) > 0 {
		log.Print("calendar: Request account sync")
		err = syncMonitor.SyncAccount(p.accountId, calendarsToSync)
		if err != nil {
			log.Print("calendar: ERROR: Fail to start account sync ", p.accountId, " message: ", err)
		}
	}

	return nil, nil
}

func (p *GCalendarPlugin) parseChangesResponse(resp *http.Response) ([]event, error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp errorResp
		log.Print("calendar: Invalid response:", errResp.Err.Code)
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
		log.Print("calendar: Fail to decode")
		return nil, err
	}

	for _, ev := range events.Events {
		log.Print("calendar: Found event: ", ev.Etag, ev.Summary)
	}

	return events.Events, nil
}

func (p *GCalendarPlugin) requestChanges(accessToken string, calendar string, lastSyncDate string) (*http.Response, error) {
	u, err := baseUrl.Parse("")
	if err != nil {
		return nil, err
	}
	u.Path += calendar + "/events"

	//GET https://www.googleapis.com/calendar/v3/calendars/<calendar>/events?showDeleted=true&singleEvents=true&updatedMin=2016-04-06T10%3A00%3A00.00Z&fields=description%2Citems(description%2Cetag%2Csummary)&key={YOUR_API_KEY}
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
