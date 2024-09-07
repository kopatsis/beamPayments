package actions

import (
	"beam_payments/redis"
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

	timeout := time.After(90 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ex, err := redis.GetQueue(id)
			if err == nil && ex {
				conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
				return nil
			}
		case <-timeout:
			conn.WriteMessage(websocket.TextMessage, []byte("error"))
			return nil
		}
	}
}
