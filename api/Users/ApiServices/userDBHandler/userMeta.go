package userDB

import(
	
	"fmt"

	"time"

	"crypto/subtle"

	"github.com/jackc/pgx"

)

type User struct{
	Name, Email string
	PassHash, Nonce []byte	
	MaxCollections int32
	Longestview time.Duration
}

// Acquires the provided user from the database with no authentication.
//
// Internal usage only to perform password authentication or acquire email.
func GetUser(pool *pgx.ConnPool, user string) (*User, error) {

	u:= User{}
	
	var LongestviewAsInt int64
	err := pool.QueryRow("getUser",
		user).Scan(&u.Name, &u.Email,
			&u.PassHash, &u.Nonce,
			&u.MaxCollections, &LongestviewAsInt)
	if err!=nil {
		return nil, errorHandle(err, ScanError)
	}

	u.Longestview = time.Duration(LongestviewAsInt)

	return &u, nil

}

// Adds a new user and returns a fresh session key.
//
// Can fail to add session key *after* adding the user, this
// is unlikely though thanks to the table constraints.
func AddUser(pool *pgx.ConnPool, user,
	email, password string) ([]byte, error) {
	
	tx, err:= pool.Begin()
	if err!=nil {
		return nil,
		fmt.Errorf("failed to grab a transaction,", err)
	}
	// Make sure we can safely exit at any time
	defer tx.Rollback()

	// Hash their password and get a complementary nonce.
	passHash, nonce, err:= derivePassword([]byte(password))
	if err!=nil {
		return nil, errorHandle(err, "failed to derive password")
	}

	// Send the user away to the db
	_, err = tx.Exec("addUser", user, email, passHash, nonce)
	if err!=nil {
		return nil, fmt.Errorf("failed to send user", err)
	}

	err = addSub(tx, user)
	if err!=nil {
		return nil, fmt.Errorf("failed to setup sub", err)
	}

	// A user needs to have a subscription as well as a meta entry.
	//
	// Failing to add either results in a broken user so we avoid that!
	tx.Commit()

	// Send a new session off to the db
	return AddSession(pool, user)

}

// Sets the password for a given user with no authentication.
// Uses a transaction to ensure atomicity
func SetPassword(tx *pgx.Tx, user, password string) error {

	// Hash their password and get a complementary nonce.
	passHash, nonce, err:= derivePassword([]byte(password))
	if err!=nil {
		return errorHandle(err, "failed to derive password")
	}

	// Send the user away to the db
	_, err = tx.Exec("setPassword", user, passHash, nonce)
	if err!=nil {
		return fmt.Errorf("failed to send fresh password", err)
	}

	return nil

}

// Sets the maximum collection count for a user with no authentication.
func SetMaxCollections(pool *pgx.ConnPool, user string, maxCollections int32) error {
	
	_, err:= pool.Exec("setMaxCollections", user, maxCollections)
	if err!=nil {
		return fmt.Errorf("failed to send new maximum", err)
	}

	return nil

}

// Authenticates a user based on a password basis
func PasswordAuthUser(pool *pgx.ConnPool,
	user, password string) (bool, error) {
	
	// Grab the user from the database if they exist
	u, err:= GetUser(pool, user)
	if err!=nil {
		return false, err
	}

	// Hash the user's provided password using scrypt
	providedHash, err:= derivePasswordWithNonce([]byte(password), u.Nonce)
	if err!=nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(u.PassHash, providedHash) == 1, nil
}

// Authenticates a user and returns a fresh session key
func Login(pool *pgx.ConnPool,
	user, password string) ([]byte, error) {

	// Make sure they are who they say they are
	valid, err:= PasswordAuthUser(pool, user, password)
	if err!=nil || !valid {
		return nil, errorHandle(err, "failed to authenticate user")
	}

	return AddSession(pool, user)

}

// Authenticates a reset request, resets the user's password, and
// delete the request used.
func ChangePassword(pool *pgx.ConnPool,
	user, password, reset string) (error) {

	// We need not validate in a transaction
	err:= ValidateReset(pool, user, reset)
	if err!=nil {
		return fmt.Errorf("failed to validate reset", err)
	}

	tx, err:= pool.Begin()
	if err!=nil {
		return fmt.Errorf("failed to grab a transaction,", err)
	}
	// Make sure we can safely exit at any time
	defer tx.Rollback()

	err = SetPassword(tx, user, password)
	if err!=nil{
		return fmt.Errorf("failed to set new password", err)
	}

	tx.Commit()

	return nil


}