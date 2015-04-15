/*
Sends a sanely derived session off to the database

Takes:
	name - string, user that owns it
	sessionKey - []byte, a valid session key
	startValid, endValid - timestamps, for preventing abuse
*/

INSERT INTO users.sessions 
(name, sessionKey, startValid, endValid) 
VALUES
($1, $2, $3, $4)