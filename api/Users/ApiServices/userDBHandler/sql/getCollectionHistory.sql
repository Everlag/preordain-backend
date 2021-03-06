/*
Acquires the total history of a user's collection.

Takes:
	owner - string, user that owns it
	collection - string, collection of that user
*/

SELECT cardName, setName, quality, quantity, comment, lang, lastUpdate
FROM
users.collectionHistory WHERE owner=$1 AND collection=$2