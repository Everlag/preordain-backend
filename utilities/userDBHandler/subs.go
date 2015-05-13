package userDB

import(

	"github.com/jackc/pgx"

	"time"

	"fmt"
)

type Subscription struct{
	Name, Plan, CustomerID, SubID string
	StartTime time.Time
}

// Adds a new subscription to a user or updates an existing one.
//
// Subscriptions have a foreign key dependency on a user existing
// so subs cannot be populated prior to user creation.
//
// A unique key provision prevents users from getting double charged
// as long as we check to ensure that we aren't setting the same twice.
func ModSub(pool *pgx.ConnPool, user, sub,
	customerID, subID string) (error) {

	s, err:= GetSub(pool, user)
	if err!=nil {
		return err
	}
	if s.Plan == sub {
		return fmt.Errorf("Duplicated sub attempt")
	}

	// Ensure that we can't change sub status without changing
	// its actual effects.
	tx, err:= pool.Begin()
	if err!=nil {
		return fmt.Errorf("failed to grab a transaction,", err)
	}
	// Make sure we can safely exit at any time
	defer tx.Rollback()

	// Send the new subscription details off to the db.
	_, err = tx.Exec("modSub", user, sub, time.Now(),
		customerID, subID)
	if err!=nil {
		return err
	}

	err = setSubEffects(tx, user, sub)
	if err!=nil {
		return err
	}

	tx.Commit()
	
	return err

}


// Acquires the subscription details of a given user.
func GetSub(pool *pgx.ConnPool, user string) (*Subscription, error) {
	
	s:= Subscription{}

	err:= pool.QueryRow("getSub", user).Scan(
		&s.Name , &s.Plan ,
		&s.CustomerID, &s.SubID,
		&s.StartTime)
	if err!=nil{
		return nil, errorHandle(err, ScanError)
	}

	return &s, nil

}

// Sets the user's current subscription into effect.
//
// Currently, this sets maxcollections and longestview for a users.meta
// entry matching this user.
func setSubEffects(tx *pgx.Tx, user, sub string) error {
	var maxCollections int
	var longestview time.Duration

	switch sub{
	case "Sensei's Top":
		maxCollections = SubTiersToCollections[sub]
		longestview = noTimeLimit
	case "Preordain":
		maxCollections = SubTiersToCollections[sub]
		longestview = noTimeLimit
	default:
		maxCollections = SubTiersToCollections[sub]
		longestview = defaultTimeLimit
	}
	
	_, err:= tx.Exec("setSubEffects", user, maxCollections, int64(longestview))

	return err
	
}

// Adds a default free subscription to a user.
//
// Use as a transaction to ensure a user can't exist without a
// subscription level
func addSub(tx *pgx.Tx, user string) error {
	_, err:= tx.Exec("modSub", user, DefaultSubLevel, time.Now(),
		"noCust", "noSub")

	return err
}