/*
Removes a provided session for a user

Takes:
	name - string, user that owns it
	sessionKey - []byte, a valid session key
*/

DELETE FROM users.sessions WHERE name=$1 AND sessionKey=$2