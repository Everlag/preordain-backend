package priceDB

import (
	"fmt"

	"time"

	"bytes"
)

// Declare a wrapper for time.Time so we can
// marshal it to a timestamp rather than a string
type Timestamp time.Time

func (t *Timestamp) MarshalJSON() ([]byte, error) {

	ts := time.Time(*t).Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil

}

var standardEM = []byte{226, 128, 148}
var oddEM = []byte{151}

// For normalization, we remove the em dash
func NormalizeEMDash(text string) string {

	text = string(bytes.Replace([]byte(text), oddEM, standardEM, -1))

	return text

}