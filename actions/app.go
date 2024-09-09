package actions

import (
	"log"
	"net/http"
	"os"
	"sync"

	"beam_payments/actions/cron"
	"beam_payments/actions/multipass"
	"beam_payments/locales"
	"beam_payments/middleware"
	"beam_payments/public"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/middleware/csrf"
	"github.com/gobuffalo/middleware/forcessl"
	"github.com/gobuffalo/middleware/i18n"
	"github.com/gobuffalo/middleware/paramlogger"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/account"
	"github.com/unrolled/secure"
)

var ENV = envy.Get("GO_ENV", "development")

var (
	app     *buffalo.App
	appOnce sync.Once
	T       *i18n.Translator
)

func App() *buffalo.App {
	appOnce.Do(func() {

		stripe.Key = os.Getenv("STRIPE_SECRET")

		acct, err := account.Get()
		if err != nil {
			log.Fatalf("Stripe API key test failed: %v", err)
		}
		log.Printf("Stripe API key test succeeded: Account ID = %s, Email = %s", acct.ID, acct.Email)

		app = buffalo.New(buffalo.Options{
			Env: ENV,
		})

		app.Use(forceSSL())

		app.Use(paramlogger.ParameterLogger)

		app.Use(csrf.New)

		app.Use(middleware.CookieMiddleware)

		cron.ScheduledTasks()

		// Wraps each request in a transaction.
		//   c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		// app.Use(popmw.Transaction(models.DB))
		// Setup and use translations:
		app.Use(translations())

		app.GET("/", GetHandler)
		app.GET("/loginerror", LoginErrorHandler)

		app.POST("/subscription", PostPayHandler)
		app.PATCH("/subscription/cancel", PostCancelHandler)
		app.PATCH("/subscription/uncancel", PostUncancelHandler)
		app.PATCH("/subscription", PostUpdatePayment)

		app.POST("/multipass", multipass.Multipass)

		app.POST("/webhook", HandleStripeWebhook)
		app.POST("/webhook/equivalent/:id", HandleEquivalentWebhook)

		app.POST("/administrative/logout", HandleLogAllOut)
		app.POST("/administrative/delete", HandleDeleteAccount)

		app.POST("/check/:id", ExternalGetHandler)

		app.POST("/helpemail", InternalEmailHandler)
		app.POST("/administrative/helpemail", ExternalEmailHandler)

		app.GET("/websocket/:id", WebSocketHandler)

		app.POST("/logout", HandleUserLogout)

		app.ServeFiles("/", http.FS(public.FS()))

	})

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(locales.FS(), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
