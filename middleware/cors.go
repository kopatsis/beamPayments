package middleware

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
)

func CORSMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {

		origin := c.Request().Header.Get("Origin")
		allowedOrigins := []string{"http://localhost:5173", "http://localhost:8080", "https://shortentrack.com", "https://api.shortentrack.com"}

		var allowed bool
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				break
			}
		}

		fmt.Println("triggered???")

		if allowed {
			c.Response().Header().Set("Access-Control-Allow-Origin", origin)
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-ID, X-Passcode-ID")
			c.Response().Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		if c.Request().Method == http.MethodOptions {
			return c.Render(204, nil)
		}

		return next(c)
	}
}
