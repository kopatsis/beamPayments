package actions

import "github.com/gobuffalo/buffalo"

func HandleLogAllOut(c buffalo.Context) error {
	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))
}

func HandleDeleteAccount(c buffalo.Context) error {
	response := map[string]any{"success": true}
	return c.Render(200, r.JSON(response))
}
