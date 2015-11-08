/*
Sends a sanely derived reset off to the database

Takes:
	name - string, user that owns it
	sessionKey - []byte, a valid session key
	startValid, endValid - timestamps, for preventing abuse
*/

INSERT INTO users.resets 
(name, resetKey, startValid, endValid) 
VALUES
($1, $2, $3, $4)