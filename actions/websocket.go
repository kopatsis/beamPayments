package actions

import (
	"beam_payments/models/badger"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebSocketHandler(c buffalo.Context) error {
	id := c.Param("id")
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	timeout := time.After(3 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if badger.GetQueue(id) {
				conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
				return nil
			}
		case <-timeout:
			conn.WriteMessage(websocket.TextMessage, []byte("error"))
			return nil
		}
	}
}
