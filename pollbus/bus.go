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

package pollbus

import (
	"fmt"
	"runtime"

	"log"

	"launchpad.net/go-dbus/v1"
)

const (
	busInterface = "com.ubuntu.AccountPolld"
	busPath      = "/com/ubuntu/AccountPolld"
	busName      = "com.ubuntu.AccountPolld"
)

type PollBus struct {
	conn     *dbus.Connection
	msgChan  chan *dbus.Message
	PollChan chan bool
}

func New(conn *dbus.Connection) *PollBus {
	p := &PollBus{
		conn:     conn,
		msgChan:  make(chan *dbus.Message),
		PollChan: make(chan bool),
	}
	runtime.SetFinalizer(p, clean)
	return p
}

func clean(p *PollBus) {
	p.conn.UnregisterObjectPath(busPath)
	close(p.msgChan)
	close(p.PollChan)
}

func (p *PollBus) Init() error {
	name := p.conn.RequestName(busName, dbus.NameFlagDoNotQueue)
	err := <-name.C
	if err != nil {
		return fmt.Errorf("bus name could not be take: %s", err)
	}

	go p.watchMethodCalls()
	p.conn.RegisterObjectPath(busPath, p.msgChan)

	return nil
}

func (p *PollBus) SignalDone() error {
	signal := dbus.NewSignalMessage(busPath, busInterface, "Done")
	if err := p.conn.Send(signal); err != nil {
		return err
	}
	return nil
}

func (p *PollBus) watchMethodCalls() {
	for msg := range p.msgChan {
		var reply *dbus.Message
		switch {
		case msg.Interface == busInterface && msg.Member == "Poll":
			log.Println("Received Poll()")
			p.PollChan <- true
			reply = dbus.NewMethodReturnMessage(msg)
		default:
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
		}
		if err := p.conn.Send(reply); err != nil {
			log.Println("Could not send reply:", err)
		}
	}
}
