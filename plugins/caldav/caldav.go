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

package caldav

import (
    "fmt"
    "bytes"
	"log"
	"net/http"
	"net/url"
	"os"
    "io/ioutil"
    "strings"
    "time"

	"launchpad.net/account-polld/accounts"
	"launchpad.net/account-polld/plugins"
    "launchpad.net/account-polld/syncmonitor"
)

const (
	APP_ID     = "com.ubuntu.calendar_calendar"
	pluginName = "caldav"
)

type CalDavPlugin struct {
	accountId uint
}

func New(accountId uint) *CalDavPlugin {
	return &CalDavPlugin{accountId: accountId}
}

func (p *CalDavPlugin) ApplicationId() plugins.ApplicationId {
	return plugins.ApplicationId(APP_ID)
}

func (p *CalDavPlugin) Poll(authData *accounts.AuthData) ([]*plugins.PushMessageBatch, error) {
	// This envvar check is to ease testing.
	if token := os.Getenv("ACCOUNT_POLLD_TOKEN_CALDAV"); token != "" {
		log.Print("Using token from: ACCOUNT_POLLD_TOKEN_CALDAV env var")
		authData.AccessToken = token
	}

	log.Print("Check calendar changes for account:", p.accountId)

	syncMonitor := syncmonitor.NewSyncMonitor()
	if syncMonitor == nil {
		log.Print("Sync monitor not available yet.")
		return nil, nil
	}

	state, err := syncMonitor.State()
	if err != nil {
		log.Print("Fail to retrieve sync monitor state ", err)
		return nil, nil
	}
	if state != "idle" {
		log.Print("Sync monitor is not on 'idle' state, try later!")
		return nil, nil
	}

	calendars, err := syncMonitor.ListCalendarsByAccount(p.accountId)
	if err != nil {
		log.Print("Calendar plugin ", p.accountId, ": cannot load calendars: ", err)
		return nil, nil
	}

	var calendarsToSync []string
	log.Print("Number of calendars for account:", p.accountId, " size:", len(calendars))

	for id, calendar := range calendars {
		lastSyncDate, err := syncMonitor.LastSyncDate(p.accountId, id)
		if err != nil {
			log.Print("\tcalendar: ", id, ", cannot load previous sync date: ", err, ". Try next time.")
			continue
		} else {
			log.Print("\tcalendar: ", id, " Url: ", calendar, " last sync date: ", lastSyncDate)
		}

		var needSync bool
		needSync = (len(lastSyncDate) == 0)

		if !needSync {
			resp, err := p.requestChanges(authData, calendar, lastSyncDate, true)
			if err != nil {
				log.Print("\tERROR: Fail to query for changes with UTC times: ", err)
				continue
			}

			needSync, err = p.containEvents(resp)
			if err != nil {
				log.Print("\tERROR: Fail to parse changes with UTC times: ", err)
				if err == plugins.ErrTokenExpired {
					log.Print("\t\tAbort poll")
					return nil, err
				} else {
					continue
				}
			}

            if !needSync {
                // WORKAROUND: Query again without convert times to UTC
                // if the server does not report the modified time in UTC we try local time
                resp, err := p.requestChanges(authData, calendar, lastSyncDate, false)
			    if err != nil {
				    log.Print("\tERROR: Fail to query for changes with local time: ", err)
				    continue
			    }

			    needSync, err = p.containEvents(resp)
			    if err != nil {
				    log.Print("\tERROR: Fail to parse changes with local time: ", err)
				    if err == plugins.ErrTokenExpired {
					    log.Print("\t\tAbort poll")
					    return nil, err
				    } else {
					    continue
				    }
			    }
            }
		}

		if needSync {
			log.Print("\tCalendar needs sync: ", id)
			calendarsToSync = append(calendarsToSync, id)
		} else {
			log.Print("\tFound no calendar updates for account: ", p.accountId, " calendar: ", id)
		}
	}

	if len(calendarsToSync) > 0 {
		log.Print("Request account sync")
		err = syncMonitor.SyncAccount(p.accountId, calendarsToSync)
		if err != nil {
			log.Print("ERROR: Fail to start account sync ", p.accountId, " message: ", err)
		}
	}

	return nil, nil
}

func (p *CalDavPlugin) containEvents(resp *http.Response) (bool, error) {
	defer resp.Body.Close()
    log.Print("RESPONSE CODE ----:", resp.StatusCode)
    
	if resp.StatusCode != 207 {
		var errResp errorResp
		log.Print("Invalid response:", errResp.Err.Code)
		return false, nil
	} else {
        data, err:= ioutil.ReadAll(resp.Body)
        if err != nil {
            return false, err
        }
        fmt.Printf("DATA: %s", data)
        return strings.Contains(string(data), "BEGIN:VEVENT"), nil
    }

    return false, nil
}

func (p *CalDavPlugin) requestChanges(authData *accounts.AuthData, calendar string, lastSyncDate string, useUTCTime bool) (*http.Response, error) {
	u, err := url.Parse(calendar)
	if err != nil {
		return nil, err
	}
    startDate, err := time.Parse(time.RFC3339, lastSyncDate)
    if err != nil {
        log.Print("Fail to parse date: ", lastSyncDate)
        return nil, err
    }

    // Start date will be one minute before last sync
    startDate = startDate.Add(time.Duration(-1)*time.Minute)

    // End Date will be one year in the future from now
    endDate := time.Now().AddDate(1,0,0)

    var dateFormat string
    if useUTCTime {
        dateFormat = "20060102T150405Z"
        endDate = endDate.UTC()
    } else {
        dateFormat = "20060102T150405"
        startDate = startDate.Local()
    }


    log.Print("Calendar Url:", calendar)
	//u.Path += "/remote.php/caldav/calendars/renatox@gmail.com/" + calendar

	//GET https://my.owndrive.com:443/remote.php/caldav/calendars/renatox%40gmail.com/teste/
    query := "<c:calendar-query xmlns:d=\"DAV:\" xmlns:c=\"urn:ietf:params:xml:ns:caldav\">\n"
    query += "<d:prop>\n"
    query +=    "<d:getetag />\n"
    query +=    "<c:calendar-data />\n"
    query += "</d:prop>\n"
    query += "<c:filter>\n"
    query +=    "<c:comp-filter name=\"VCALENDAR\">\n"
    query +=    "<c:comp-filter name=\"VEVENT\">\n"
    query +=    "<c:prop-filter name=\"LAST-MODIFIED\">\n"
    query +=        "<c:time-range start=\"" + startDate.Format(dateFormat) + "\" end=\"" + endDate.Format(dateFormat) + "\"/>\n"
    query +=    "</c:prop-filter>\n"
    query +=    "</c:comp-filter>\n"
    query +=    "</c:comp-filter>\n"
    query += "</c:filter>\n"
    query += "</c:calendar-query>\n"
    log.Print("Query: ", query)
	req, err := http.NewRequest("REPORT", u.String(), bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}
    req.Header.Set("Depth", "1");
    req.Header.Set("Prefer", "return-minimal");
    req.Header.Set("Content-Type", "application/xml; charset=utf-8");
    req.SetBasicAuth(authData.UserName, authData.Secret)

	return http.DefaultClient.Do(req)
}
