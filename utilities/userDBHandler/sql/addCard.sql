/*

UPSERTs into userCollectionContents using the add_card function.

Could potentially loop forever but good god does postgres need this natively.

add_card has the format
	add_card(specOwner TEXT, specCollection TEXT,
			specCardName TEXT, specSetName TEXT, specComment TEXT,
			specQuantity int, specQuality possibleQuality, specTime timestamp)

Takes:
	owner - string, user that owns it
	collection - string, collection of that user
	cardName - string, the mtg card
	setName - string, the mtg set
	comment - string, a user comment
	quality - string, a defined quality
	lang - string, a language in mtg
	quantity - int, how many cards
*/

SELECT add_card($1, $2, $3, $4, $5, $6, $7, $8, $9);