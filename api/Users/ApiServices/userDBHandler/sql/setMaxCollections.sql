/*
Updates a user's password on the database

Takes:
	name - string, user that owns it
	maxCollections - int32, the most collections that user can have
*/

UPDATE users.meta
SET maxCollections = $2
WHERE
name=$1