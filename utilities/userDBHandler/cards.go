package userDB

import(

	"fmt"
	"time"

	"github.com/jackc/pgx"

)

type Card struct{
	Name, Set, Quality, Comment, Lang string
	Quantity int32
	LastUpdate time.Time
}

// Safely adds a card using a transaction to apply to
// both the current status and history.
func AddCard(pool *pgx.ConnPool, sessionKey []byte,
	user, collection,
	Name, Set, Comment, Quality, Lang string,
	Quantity int32, LastUpdate time.Time) error {

	// Start the transaction
	tx, err:= pool.Begin()
	if err!=nil {
		return fmt.Errorf("failed to grab a transaction,", err)
	}
	// Make sure we can safely exit at any time
	defer tx.Rollback()

	// Make sure the user's collection exists
	coll, err:= GetCollectionMeta(pool, sessionKey, user, collection)
	if err!=nil {
		return fmt.Errorf("failed to check collection exists")
	}
	if coll.Name != collection {
		return fmt.Errorf("no such collection exists")
	}


	err = insertCard(tx,
		user, collection,
		Name, Set, Comment,
		Quantity,
		Quality, Lang,
		LastUpdate)
	if err!=nil {
		return fmt.Errorf("failed to add to history, ", err)
	}

	tx.Commit()

	return nil

}

// Safely adds a number of cards using a transaction to apply to
// both the current status and history.
//
// Inserting multiple cards per single transaction is a lot more
// efficient and should be the aim.
func AddCards(pool *pgx.ConnPool, sessionKey []byte,
	user, collection string,
	cards []Card) error {
	
	// Start the transaction
	tx, err:= pool.Begin()
	if err!=nil {
		return fmt.Errorf("failed to grab a transaction,", err)
	}
	// Make sure we can safely exit at any time
	defer tx.Rollback()

	// Make sure the user's collection exists
	coll, err:= GetCollectionMeta(pool, sessionKey, user, collection)
	if err!=nil {
		return fmt.Errorf("failed to ensure collection exists")
	}
	if coll.Name != collection {
		return fmt.Errorf("no such collection exists")
	}

	for _, aCard:= range cards{

		err:= insertCard(tx,
						user, collection,
						aCard.Name, aCard.Set, aCard.Comment,
						aCard.Quantity, aCard.Lang, aCard.Quality, aCard.LastUpdate)

		if err!=nil {
			return fmt.Errorf("failed to insert card", err)
		}
	}

	tx.Commit()

	return nil

}

// Inserts a card into the db using a passed transaction
func insertCard(tx *pgx.Tx,
	user, collection,
	Name, Set, Comment string,
	Quantity int32, Lang string,
	Quality string, LastUpdate time.Time) error {

	var err error

	// Send the contents upsert
	_, err = tx.Exec("addCard",
					user, collection,
					Name, Set, Comment,
					Quantity, Quality, Lang,
					LastUpdate)
	if err!=nil {
		return fmt.Errorf("failed to add to contents, ", err)
	}

	// Send the historical row
	_, err = tx.Exec("addCardHistorical",
					user, collection,
					Name, Set, Comment,
					Quantity, Quality, Lang,
					LastUpdate)
	if err!=nil {
		return fmt.Errorf("failed to add to history, ", err)
	}

	return nil
	
}

// Acquires every change to a specified user's collection
func GetCollectionHistory(pool *pgx.ConnPool, sessionKey []byte,
	user, collection string) ([]Card, error) {
	
	var err error

	// Authenticate the request
	if sessionKey != nil {
		err = SessionAuth(pool, user, sessionKey)
		if err!=nil{
			return nil, errorHandle(err, "authorization Failed, invalid session key")
		}	
	}

	// Grab everything and pack it nicely to be returned
	rows, err := pool.Query("getCollectionHistory", user, collection)
	if err!=nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next(){
		c:= Card{}
		err = rows.Scan(&c.Name, &c.Set,
			&c.Quality, &c.Quantity,
			&c.Comment, &c.Lang, &c.LastUpdate)
		if err!=nil {
			return nil, errorHandle(err, ScanError)
		}

		cards = append(cards, c)
	}

	return cards, nil

}

// Acquires all cards in a specified user's collection
func GetCollectionContents(pool *pgx.ConnPool, sessionKey []byte,
	user, collection string) ([]Card, error) {
	
	// Authenticate the request
	if sessionKey != nil {
		err:= SessionAuth(pool, user, sessionKey)
		if err!=nil{
			return nil, errorHandle(err, "authorization Failed, invalid session key")
		}	
	}
	
	// Grab everything and pack it nicely to be returned
	rows, err := pool.Query("getCollectionContents", user, collection)
	if err!=nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next(){
		c:= Card{}
		err = rows.Scan(&c.Name, &c.Set,
			&c.Quality, &c.Quantity,
			&c.Comment, &c.Lang, &c.LastUpdate)
		if err!=nil {
			return nil, errorHandle(err, ScanError)
		}

		cards = append(cards, c)
	}

	return cards, nil
	
}