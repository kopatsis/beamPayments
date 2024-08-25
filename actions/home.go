package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

func HomeHandler(c buffalo.Context) error {

	// err := bdb.DB.Update(func(txn *badger.Txn) error {
	// 	err := txn.Set([]byte("session-id"), []byte("cookie-data"))
	// 	return err
	// })

	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}
