/*
Acquires the every session key for a provided user that matches
the provided session key and is valid.

The fact that a row is returned means that the provided session/user
combo is valid.

Takes:
	name - string, user that owns it
	sessionKey - []byte, a valid session key
*/

SELECT name, sessionKey, startValid, endValid
FROM
users.sessions
WHERE name=$1 AND sessionKey=$2 AND endValid > now()