package sendgrid

import (
	"os"

	"github.com/sendgrid/sendgrid-go"
)

var SGClient *sendgrid.Client

func init() {
    SGClient = sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
}
