/*
Acquires a user from the server with no authentication

Takes:
	name - string, user that owns it
*/

SELECT name, email, passhash, nonce, maxcollections
FROM
users.meta WHERE name=$1