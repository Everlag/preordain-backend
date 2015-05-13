/*
Updates a user on the database to have the desired subscription facets

Takes:
	name - string, user that owns it
	maxcollections - int, how many collections that user may have
	longestview - timestamp, how far back into the future a user may view.

*/

UPDATE users.meta
SET maxcollections = $2, longestview = $3
WHERE name=$1