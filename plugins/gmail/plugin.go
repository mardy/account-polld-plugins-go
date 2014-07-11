/*
 Copyright 2014 Canonical Ltd.
 Authors: Sergio Schvezov <sergio.schvezov@canonical.com>

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

package gmail

import "launchpad.net/account-polld/plugins"

const APP_ID = "com.ubuntu.developer.webapps.webapp-gmail"

type GmailPlugin struct {
}

func New() *GmailPlugin {
	return &GmailPlugin{}
}

func (p *GmailPlugin) Poll(tokens plugins.AuthTokens) (*[]plugins.Notification, error) {
	return nil, nil
}

func (p *GmailPlugin) ApplicationId() plugins.ApplicationId {
	return plugins.ApplicationId(APP_ID)
}
