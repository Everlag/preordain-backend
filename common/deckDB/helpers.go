package deckDB

import (
	"fmt"

	"time"
)

// Declare a wrapper for time.Time so we can
// marshal it to a timestamp rather than a string
type Timestamp time.Time

func (t Timestamp) MarshalJSON() ([]byte, error) {

	ts := time.Time(t).Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil

}