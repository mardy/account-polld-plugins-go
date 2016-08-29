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

package plugins

import (
	"log"
)

type PluginRunner struct {
	watcher          *Ipc
	plugin           Plugin
	postWatch        chan *PostWatch
	authChan         chan AuthData
	penaltyCount     int
	authFailureCount int
}

type PostWatch struct {
	appId   ApplicationId
	batches []*PushMessageBatch
}

func NewPluginRunner(plugin Plugin) *PluginRunner {
	authChan := make(chan AuthData, 1)
	return &PluginRunner{
		watcher:   NewIpc(authChan),
		plugin:    plugin,
		postWatch: make(chan *PostWatch),
		authChan:  authChan,
	}
}

func (r *PluginRunner) Delete() {
	close(r.authChan)
}

func (r *PluginRunner) Run() {
	go r.watcher.Run()
	for {
		select {
		case data := <-r.authChan:
			log.Println("Got data, access token is ", data.AccessToken)
			r.poll(&data)
		case post := <-r.postWatch:
			log.Println("Got reply")
			r.watcher.PostMessages(post.batches)
		}
	}
}

func (r *PluginRunner) poll(authData *AuthData) error {
	log.Println("Polling account", authData.AccountId)

	if bs, err := r.plugin.Poll(authData); err != nil {
		log.Print("Error while polling ", authData.AccountId, ": ", err)
		return err
	} else {
		for _, b := range bs {
			log.Println("Account", authData.AccountId, "has", len(b.Messages), b.Tag, "updates to report")
		}
		r.postWatch <- &PostWatch{batches: bs, appId: r.plugin.ApplicationId()}
		return err
	}
}
