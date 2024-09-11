package actions

import (
	"beam_payments/models"
	"beam_payments/redis"
	"context"
	"net/http"
	"strings"
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
	start := time.Now()

	id := c.Param("id")
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	pubsub := redis.RDB.Subscribe(context.Background(), "Subscription")
	defer pubsub.Close()

	timeoutDuration := 180 * time.Second
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-pubsub.Channel():
			parts := strings.Split(msg.Payload, " --- ")
			if len(parts) == 2 {
				subscriptionID, status := parts[0], parts[1]
				if subscriptionID == id && status == "Success" {
					conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
					return nil
				} else if subscriptionID == id && status == "Fail" {
					conn.WriteMessage(websocket.TextMessage, []byte("error"))
					return nil
				}
			}
		case <-ticker.C:
			sub, none, err := models.GetSubscriptionBySubID(id)
			if none || err != nil || !sub.Processing {
				conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
				return nil
			}
		case <-time.After(timeoutDuration - time.Since(start)):
			conn.WriteMessage(websocket.TextMessage, []byte("timeout"))
			return nil
		}
	}
}
