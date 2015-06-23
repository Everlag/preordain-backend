package userDB

import(
	
	"fmt"
	"time"

	"crypto/subtle"
	"crypto/sha256"

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

	// Store the hash of the key so read-only exploits are ineffective
	hashed:= sha256.Sum256(key)
	
	// Setup a session to send to the db
	now:= time.Now()
	freshSession:= Session{
		Name: user,
		SessionKey: hashed[:],
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
	
	// Hash the key so we compare hashes instead of contents
	hashed:= sha256.Sum256(sessionKey)

	rows, err := pool.Query("getSessions", user, hashed[:])
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
		subtle.ConstantTimeCompare(hashed[:], s.SessionKey) == 1 &&
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

// Generates a request reset by inserting a reset key valid for this user
func RequestReset(pool *pgx.ConnPool, user string) (string, error) {

	// Ensure the user hasn't received a reset code within
	// the possible duration of another valid reset code
	now:= time.Now()
	resets, err:= getAllResets(pool, user)
	if err!=nil {
		return "", errorHandle(err, "failed to acquire all resets")
	}

	for _, r:= range resets{
		// A reset code valid after the now + resetValidTime
		// means we don't send another token.
		delta:= now.Sub(r.EndValid)
		if delta < resetValidTime {
			return "", fmt.Errorf("requested reset when valid already exists")
		}
	}

	// Acquire a fresh session key of length 256 bits
	key:= randString(ResetLength)
	// Store the hash of the key and use that for comparisons
	hashed:= sha256.Sum256([]byte(key))
	
	// Setup a session to send to the db
	freshReset:= Reset{
		Name: user,
		ResetKey: hashed[:],
		StartValid: now,
		EndValid: now.Add(resetValidTime),
	}

	// Send the session off
	err = SendReset(pool, freshReset)
	if err!=nil {
		return "", errorHandle(err, "failed to send fresh reset off to db")
	}

	return key, err

}

func ValidateReset(pool *pgx.ConnPool, user, resetKey string) error {
	
	// Request a hash matching the key's hash.
	hashed:= sha256.Sum256([]byte(resetKey))

	rows, err := pool.Query("getReset", user, hashed[:])
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
		subtle.ConstantTimeCompare(hashed[:],
			[]byte(r.ResetKey)) == 1 &&
		now.Before(r.EndValid) && now.After(r.StartValid) {
			return nil
		}
	}

	return fmt.Errorf("invalid Authentication")

}

// Acquires all valid resets for a given user
func getAllResets(pool *pgx.ConnPool, user string) ([]Reset, error) {

	rows, err := pool.Query("getAllResets", user)
	if err!=nil {
		return nil, err
	}
	defer rows.Close()

	resets:= make([]Reset, 0)

	for rows.Next(){
		r:= Reset{}
		err = rows.Scan(&r.Name, &r.ResetKey,
			&r.StartValid, &r.EndValid)
		if err!=nil {
			return nil, errorHandle(err, ScanError)
		}

		resets = append(resets, r)
	}

	return resets, nil
}