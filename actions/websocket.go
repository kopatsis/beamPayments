package actions

import (
	"beam_payments/redis"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gorilla/websocket"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/sub"
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
	ticker := time.NewTicker(9 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-pubsub.Channel():
			parts := strings.Split(msg.Payload, " --- ")
			if len(parts) == 2 {
				subscriptionID, status := parts[0], parts[1]
				if subscriptionID == id && status == "Success" {
					return conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
				} else if subscriptionID == id && status == "Fail" {
					return conn.WriteMessage(websocket.TextMessage, []byte("error"))
				}
			}
		case <-ticker.C:
			s, err := sub.Get(id, nil)
			if err == nil && (s.Status == stripe.SubscriptionStatusActive || s.Status == stripe.SubscriptionStatusPastDue) {
				return conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
			}
		case <-time.After(timeoutDuration - time.Since(start)):
			conn.WriteMessage(websocket.TextMessage, []byte("timeout"))
			return nil
		}
	}
}
