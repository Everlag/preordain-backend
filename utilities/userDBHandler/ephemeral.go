package userDB

import(
	
	"fmt"
	"time"

	"crypto/subtle"

	"github.com/jackc/pgx"

)

// How many characters a reset should be.
// len(alphanum)^ResetLength should be infeasible
// to guess within an expiry period.
const ResetLength int = 20

type Session struct{
	Name string
	SessionKey []byte	
	StartValid, EndValid time.Time
}

// Commits a provided session off to the postgres backend
func SendSession(pool *pgx.ConnPool, session Session) error {
	
	_, err:= pool.Exec("addSession",
					session.Name, session.SessionKey,
					session.StartValid, session.EndValid)

	return err

}

// Generates a fresh session key for the provided user and sends that.
//
// Returns a valid session key
func AddSession(pool *pgx.ConnPool, user string) ([]byte, error) {

	// Acquire a fresh session key of length 256 bits
	key, err:= getArrayOfRandBytes(32)
	if err!=nil {
		return nil, fmt.Errorf("failed to derive new session key, ", err)
	}
	
	// Setup a session to send to the db
	now:= time.Now()
	freshSession:= Session{
		Name: user,
		SessionKey: key,
		StartValid: now,
		EndValid: now.Add(sessionValidTime),
	}

	// Send the session off
	err = SendSession(pool, freshSession)
	if err!=nil {
		return nil, errorHandle(err, "failed to send fresh session off to db")
	}

	return key, err

}

// Authenticates a user based on the presence of a session key-name
// pair existing on the database that is valid.
//
// Constant time relative to the number of session keys on the user
func SessionAuth(pool *pgx.ConnPool, user string, 
	sessionKey []byte) error {
	
	rows, err := pool.Query("getSessions", user, sessionKey)
	if err!=nil {
		return err
	}
	defer rows.Close()

	now:= time.Now()
	for rows.Next(){
		s:= Session{}
		err = rows.Scan(&s.Name, &s.SessionKey,
			&s.StartValid, &s.EndValid)
		if err!=nil {
			return errorHandle(err, ScanError)
		}

		// Perform validation
		if s.Name == user &&
		subtle.ConstantTimeCompare(sessionKey, s.SessionKey) == 1 &&
		now.Before(s.EndValid) && now.After(s.StartValid) {
			return nil
		}
	}

	return fmt.Errorf("invalid Authentication")

}

// Remove an existing session.
func Logout(pool *pgx.ConnPool, user string, 
	sessionKey []byte) error {
	
	_, err:= pool.Exec("removeSession",
					user, sessionKey)

	return err

}

type Reset struct{
	Name string
	ResetKey []byte
	StartValid, EndValid time.Time
}

// Commits a provided reset off to the postgres backend
func SendReset(pool *pgx.ConnPool, reset Reset) error {
	
	_, err:= pool.Exec("addReset",
					reset.Name, reset.ResetKey,
					reset.StartValid, reset.EndValid)

	return err

}

// Generates a request reset by inserting a reset key valid for
func RequestReset(pool *pgx.ConnPool, user string) (string, error) {
	
	// Acquire a fresh session key of length 256 bits
	key:= randString(ResetLength)
	
	// Setup a session to send to the db
	now:= time.Now()
	freshReset:= Reset{
		Name: user,
		ResetKey: []byte(key),
		StartValid: now,
		EndValid: now.Add(resetValidTime),
	}

	// Send the session off
	err:= SendReset(pool, freshReset)
	if err!=nil {
		return "", errorHandle(err, "failed to send fresh reset off to db")
	}

	return key, err

}

func ValidateReset(pool *pgx.ConnPool, user, resetKey string) error {
	
	rows, err := pool.Query("getReset", user, resetKey)
	if err!=nil {
		return err
	}
	defer rows.Close()

	now:= time.Now()
	for rows.Next(){
		r:= Reset{}
		err = rows.Scan(&r.Name, &r.ResetKey,
			&r.StartValid, &r.EndValid)
		if err!=nil {
			return errorHandle(err, ScanError)
		}

		// Perform validation
		if r.Name == user &&
		subtle.ConstantTimeCompare([]byte(resetKey),
			[]byte(r.ResetKey)) == 1 &&
		now.Before(r.EndValid) && now.After(r.StartValid) {
			return nil
		}
	}

	return fmt.Errorf("invalid Authentication")

}