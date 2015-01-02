package UserStructs

import(

	"fmt"

	"time"

)

// If a user's session token is within userSessionExtensionLength of its expiry,
// it being used will increment it by the userSessionExtensionLength
const userSessionExtensionLength int64 = 3600 * 24 * 7

// If an action would be attempted that could be considered abusive, we rate limit
// it to prevent that abuse
const userAbusePreventionLength int64 = 3600

//a user session contains the timestamp for which it is valid up to and the key
//that it is associated with.
//
//everytime a session is used its end time should be incremented by
//userSessionExtensionLength
type UserSession struct {
	Key string
	Start int64
	End int64
}

//extends a user session if it is valid and is within
//userSessionExtensionLength of its expiry
//
//returns an error if the token is expired, otherwise returns nil
func (aSession *UserSession) extend() error {

	if !aSession.valid() {
		return fmt.Errorf("Token is expired")
	}

	now := time.Now().UTC().Unix()

	if aSession.End+userSessionExtensionLength > now {
		aSession.End += userSessionExtensionLength
	}

	return nil

}

//invalidates this session
func (aSession *UserSession) end() {

	aSession.End = -1

}

//checks if the session is past its end time.
func (aSession *UserSession) valid() bool {

	now := time.Now().UTC().Unix()
	if now > aSession.End {
		return false
	}

	return true

}

// Checks to see if the token has been issued inside the userAbusePreventionLength
// window.
func (aSession *UserSession) insideAbuseWindow() bool {
	
	now := time.Now().UTC().Unix()

	if now - aSession.Start < userAbusePreventionLength{
		return true
	}

	return false

}

//returns a fresh session valid till userSessionExtensionLength * 2 + NOW
func newSession() UserSession {
	startTime:= time.Now().UTC().Unix()
	endTime := (userSessionExtensionLength * 2) + time.Now().UTC().Unix()

	key := randString(32)

	return UserSession{
		Key: key,
		End: endTime,
		Start: startTime,
	}
}
