package badger

import (
	"time"

	"github.com/gofrs/uuid"
)

type Cookie struct {
	Banned    bool
	Passcode  uuid.UUID
	ResetDate time.Time
}

func (c *Cookie) MarshalBinary() []byte {
	return c.Passcode.Bytes()
}
