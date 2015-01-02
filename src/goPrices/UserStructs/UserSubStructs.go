//contains the data structures associated with user data as well as methods
//for handling the addition of data.
//
//trade records are immutable apart from their status as revoked.
package UserStructs

import (
	"fmt"
)



// We keep the modify history to allow for the implementation
// of a timemachine collection viewer feature and
// trade history feature.
//
// A state of the collection is not maintained as clients are expected to
// derive the state to whatever point they desire from the trade data.
type Collection struct {
	Name string

	//we keep a list of additions to the Collection.
	ModifyHistory []Trade

	//we have the current state precomputed so that clients can avoid
	//having to compute from ModifyHistory the most useful state
	//
	//contents is kept in the form [SetName][CardName]instance
	Contents map[string]*map[string]OwnedCard

	//we have permission bytes per user collection
	//Public Comments is dependent on Public History which is Dependent on
	//Public Viewing
	//allows the public to view this
	//collection without historical trades
	PublicViewing bool

	//allows the public to view this collection
	//with all the knowledge the user would have
	PublicHistory bool

	//allow the public to view this collection with comments on each trade
	PublicComments bool
}

// Sets basic permissions for a collection after performing simple sanity checks
func (aColl *Collection) SetPermissions(Viewing, History, Comments bool) (error) {
	
	if (!Viewing && History) || (!Viewing && Comments) || (!History && Comments) {
		return fmt.Errorf("Non-sane permissions requested")
	}

	aColl.PublicViewing = Viewing
	aColl.PublicHistory = History
	aColl.PublicComments = Comments

	return nil

}

// Returns a blank collection
func CreateCollection(aCollName string) Collection {
	aColl := Collection{}
	aColl.Name = aCollName
	aColl.ModifyHistory = make([]Trade, 0)
	aColl.Contents = make(map[string]*map[string]OwnedCard)
	return aColl
}

// Strips all private data away as determined by the collection's
// public accessibility fields
func (aColl *Collection) StripToPublic() (*Collection, error) {

	//create a copy of the public struct we'll return
	var stripped Collection

	stripped = Collection{
		Name:           aColl.Name,
		Contents:       aColl.Contents,
		ModifyHistory:  aColl.ModifyHistory,
		PublicViewing:  aColl.PublicViewing,
		PublicHistory:  aColl.PublicHistory,
		PublicComments: aColl.PublicComments,
	}

	if !aColl.PublicViewing {
		return nil, fmt.Errorf("Collection doesn't exist")
	} else if !aColl.PublicHistory {
		stripped.ModifyHistory = nil
	} else if !aColl.PublicComments {
		for i, aTrade := range stripped.ModifyHistory {
			cleanedTrade := aTrade
			cleanedTrade.Comment = ""
			stripped.ModifyHistory[i] = cleanedTrade
		}
	}

	return &stripped, nil

}

// Adds a preformed trade to the specified collection
//
// This is the only method with which a collection may be modified
func (aColl *Collection) AddTrade(aTrade Trade) error {

	//truncate the associated comment if necessary
	//
	//we need not worry about sanitizing the comment, clients are expected to
	//practice sane XSS prevention.
	if len(aTrade.Comment) > MaxTradeCommentLength {
		aTrade.Comment = aTrade.Comment[:MaxTradeCommentLength-4] + "..."
	}

	// Ensure this is a clean and reasonable trade before committing any state
	err:= aTrade.valid()
	if err!=nil {
		return err
	}

	//add the cards in the trade
	for _, aCard := range aTrade.Transaction {
		aColl.addCard(aCard)
	}

	//append the trade to the collection's modify history
	aColl.ModifyHistory = append(aColl.ModifyHistory, aTrade)

	return nil

}

// Adds a card to the collection. This is done without history and is not safe
// for external usage as clients are dependant on the precomputed contents being
// synchronized with the history
func (aColl *Collection) addCard(aCard OwnedCard) {
	//grab the set, create if non-existent
	setContainer, setExists := aColl.Contents[aCard.Set]
	if !setExists {
		freshSetContainer := make(map[string]OwnedCard)
		aColl.Contents[aCard.Set] = &freshSetContainer

		setContainer = &freshSetContainer
	}

	//acquire a direct way to address the set
	directSetContainer := *setContainer

	//set the card data
	cardData, cardExists := directSetContainer[aCard.Name]
	if cardExists {
		aCard.Quantity = aCard.Quantity +
			cardData.Quantity
	}

	directSetContainer[aCard.Name] = aCard
}

// The absolute maximum length of a trade comment is 140 characters.
//
// The same size as a tweet, it forces concise comments while drawing
// a connection between twitter's mindset and this.
const MaxTradeCommentLength int = 140

// Details a specific transaction to the collection database
// we have defined
type Trade struct {
	//additions and subtractions presented by trade are found in the
	//ownedCard[i].Quantity category. negatives are removals with
	//positives as additions
	Transaction []OwnedCard

	//a trade may or may not have a comment which briefly describes
	//the mindset of the trade, the reasons, or
	Comment string

	//when each transaction happens.
	//
	//each is associated with a timestamp
	TimeStamp int64

	//a bool to determine whether or not a given trade
	//has been revoked
	Revoked bool
}

// Returns an error if a card in the trade contains an invalid language
// or if the comment is too long.
func (aTrade *Trade) valid() error {
	
	if len(aTrade.Comment) > MaxTradeCommentLength {
		return fmt.Errorf("Comment for trade is too long")
	}

	for _, aCard:= range aTrade.Transaction{
		if !isSupportedLanguage(aCard.Language) {
			return fmt.Errorf("Invalid Card Language, ", aCard.Language)
		}
	}

	return nil

}

// Imports an existing trade.
//
// Truncates comments to their maximum length.
//
// Ensures language is a language Magic supports
func CreateExistingTrade(someCards []OwnedCard, TimeStamp int64,
	comment string) (Trade, error) {
	if len(comment) > MaxTradeCommentLength {
		comment = comment[:MaxTradeCommentLength-4] + "..."
	}

	aTrade:= Trade{someCards, comment, TimeStamp, false}

	return aTrade, aTrade.valid()
}

// A basic card structure to work with.
//
// Language is valid country code.
//
// Signed is the subset of cards with signatures in the Quantity.
type OwnedCard struct {
	Name     string
	Set      string
	Quantity int32

	Language string

	Signed   int32
}

func CreateCard(Name, Set string, Quantity,
	Signed int, Language string) OwnedCard {
	return OwnedCard{Name, Set,
		int32(Quantity), Language,
		int32(Signed)}
}
