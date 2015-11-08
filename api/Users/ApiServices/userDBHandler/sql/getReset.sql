/*
Acquires the every reset key for a provided user that matches
the provided reset key and is valid.

The fact that a row is returned means that the provided session/user
combo is valid. We still check in the middleware though.

Takes:
	name - string, user that owns it
	sessionKey - []byte, a valid session key
*/

SELECT name, resetKey, startValid, endValid
FROM users.resets
WHERE name=$1 AND resetKey=$2 AND endValid > now()