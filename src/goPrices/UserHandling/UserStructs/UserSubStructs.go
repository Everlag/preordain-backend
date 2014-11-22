//contains the data structures associated with user data as well as methods
//for handling the addition of data.
//
//trade records are immutable apart from their status as revoked.
package UserStructs

//we keep the modify history to allow for the implementation
//of a timemachine collection viewer feature and
//trade history feature.
//
//A state of the collection is not maintained as clients are expected to
//derive the state to whatever point they desire from the trade data.
type Collection struct {
	Name string

	//we keep a list of additions to the Collection. 
	ModifyHistory []Trade

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

	//if the collection is a deck, essentially an extended collection,
	//we must be capable marking that
	IsDeck bool
}

func CreateCollection(aCollName string) Collection {
	aColl := Collection{}
	aColl.Name = aCollName
	aColl.ModifyHistory = make([]Trade, 0)
	return aColl
}

//adds a preformed trade to the specified collection
//
//this is the only method with which a collection may be modified
func (aColl *Collection) AddExistingTrade(aTrade Trade) {

	//truncate the associated comment if necessary
	//
	//we need not worry about sanitizing the comment, clients are expected to
	//practice sane XSS prevention.
	if len(aTrade.Comment) > MaxTradeCommentLength {
		aTrade.Comment = aTrade.Comment[:MaxTradeCommentLength-4] + "..."
	}

	//append the trade to the collection's modify history
	aColl.ModifyHistory = append(aColl.ModifyHistory, aTrade)

}

//the absolute maximum length of a trade comment is 140 characters.
//
//The same size as a tweet, it forces concise comments while drawing
//a connection between twitter's mindset and this.
const MaxTradeCommentLength int = 140

//details a specific transaction to the collection database
//we have defined
type Trade struct{
	//additions and subtractions presented by trade are found in the
	//ownedCard[i].Quantity category. negatives are removals with
	//positives as additions
	Transactions []OwnedCard

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

//imports an existing trade.
//
//truncates comments to their maximum length
func CreateExistingTrade(someCards []OwnedCard, TimeStamp int64, comment string) Trade {
	if len(comment)>MaxTradeCommentLength {
		comment = comment[:MaxTradeCommentLength-4] + "..."
	}
	return Trade{someCards, comment, TimeStamp, false}
}


//A basic card structure to work with.
//
//Language is a map of languages present to amount of cards per language.
//these are a subset under quantity which is the ABSOLUTE quantity of cards.
//
//Signed is the subset of cards with signatures in the Quantity.
type OwnedCard struct {
	Name           string
	Set            string
	Quantity       int32

	Language       map[string]int32
	Signed         int32
}

func CreateCard(Name, Set string, Quantity, 
	Signed int, LanguageToCount map[string]int32) OwnedCard {
	return OwnedCard{Name, Set,
					int32(Quantity), LanguageToCount,
					int32(Signed)}
}