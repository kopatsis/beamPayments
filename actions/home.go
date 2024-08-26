package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

func HomeHandler(c buffalo.Context) error {

	return c.Render(http.StatusOK, r.HTML("home/error.plush.html"))
}
