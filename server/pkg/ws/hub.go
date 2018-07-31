// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"github.com/mateuszdyminski/auto/ingress/model"
)

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	Connections map[*Connection]bool

	// Inbound messages from the connections.
	Broadcast chan *model.FlightCrash

	// Register requests from the connections.
	Register chan *Connection

	// Unregister requests from connections.
	Unregister chan *Connection
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan *model.FlightCrash),
		Register:    make(chan *Connection),
		Unregister:  make(chan *Connection),
		Connections: make(map[*Connection]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.Connections[c] = true
		case c := <-h.Unregister:
			if _, ok := h.Connections[c]; ok {
				delete(h.Connections, c)
				close(c.Send)
			}
		case m := <-h.Broadcast:
			for c := range h.Connections {
				select {
				case c.Send <- m:
				default:
					close(c.Send)
					delete(h.Connections, c)
				}
			}
		}
	}
}
