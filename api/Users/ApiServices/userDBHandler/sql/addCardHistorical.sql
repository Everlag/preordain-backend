/*
Inserts a row into the userCollectionContents.

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

INSERT INTO users.collectionHistory 
(owner, collection, cardName, setName, comment, quantity, quality, lang, lastUpdate) 
VALUES
($1, $2, $3, $4, $5, $6, $7, $8, $9)