package util

import (
	"github.com/satori/go.uuid"
)

// UUID allocates an unique object ID.
func UUID() []byte { return uuid.NewV4().Bytes() }
