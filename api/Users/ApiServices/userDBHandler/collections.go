package userDB

import(

	"time"

	"fmt"

	"github.com/jackc/pgx"

)

type Collection struct{
	Name, Owner string
	LastUpdate time.Time
	Privacy string
}

// Commits a new collection to the database only if the user has less than
// their maximum number of collections!
func AddCollection(pool *pgx.ConnPool, sessionKey []byte,
	user, collection string) error {
	
	// Authenticate the request
	err:= SessionAuth(pool, user, sessionKey)
	if err!=nil{
		return errorHandle(err, "authorization Failed, invalid session key")
	}

	// Find how many collections we can have
	userDetails, err:= GetUser(pool, user)
	if err!=nil {
		return errorHandle(err, "failed to fetch user")
	}
	collections, err:= GetCollectionList(pool, user)
	if err!=nil {
		return errorHandle(err, "failed to fetch collection list")
	}

	if int(userDetails.MaxCollections) < (len(collections) + 1) {
		return fmt.Errorf("collection limit reached")
	}

	// Find how many collections we have

	_, err = pool.Exec("addCollection",
					user, collection)

	return err


}

// Commits new public viewing permissions to the database.
func SetCollectionPrivacy(pool *pgx.ConnPool, sessionKey []byte,
	user, collection, Privacy string) error {
	
	// Authenticate the request
	err:= SessionAuth(pool, user, sessionKey)
	if err!=nil{
		return errorHandle(err, "authorization Failed, invalid session key")
	}

	_, err = pool.Exec("setCollectionPermissions",
					user, collection, Privacy)

	return err


}

// Acquires metadata for a given collection
func GetCollectionMeta(pool *pgx.ConnPool, sessionKey []byte,
	user, collection string) (*Collection, error) {
	
	var err error

	// Authenticate the request
	if sessionKey != nil {
		err = SessionAuth(pool, user, sessionKey)
		if err!=nil{
			return nil, errorHandle(err, "authorization Failed, invalid session key")
		}	
	}

	c:= Collection{}
	
	err = pool.QueryRow("getCollectionMeta",
		user, collection).Scan(&c.Name, &c.Owner,
			&c.LastUpdate,
			&c.Privacy)
	if err!=nil {
		return nil, errorHandle(err, ScanError)
	}

	return &c, nil

}

// Acquire metadata for all collections for a given user.
func GetCollectionList(pool *pgx.ConnPool, user string) ([]Collection, error) {

	rows, err := pool.Query("getCollectionList", user)
	if err!=nil {
		return nil, err
	}
	defer rows.Close()

	var collections []Collection
	for rows.Next(){
		c:= Collection{}
		err = rows.Scan(&c.Name, &c.Privacy)
		if err!=nil {
			return nil, errorHandle(err, ScanError)
		}

		collections = append(collections, c)
	}

	return collections, nil

}