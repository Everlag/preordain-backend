/*
Acquires the latest state of a card in a user's collection.

Takes:
	owner - string, user that owns it
	collection - string, collection of that user
	cardName - string, the mtg card
	setName - string, the mtg set
	quality - string, one of the qualities we have set as valid for cards
*/

SELECT cardName, setName, quality, quantity, lastUpdate
FROM 
users.collectionContents
WHERE 
owner=$1 AND collection=$2 AND cardName=$3 AND setName=$4 AND quality=$5