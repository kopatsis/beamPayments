package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

func LoginErrorHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("errors/loginerror.plush.html"))
}
