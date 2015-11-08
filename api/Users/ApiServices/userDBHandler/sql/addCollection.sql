/*
Sends a fresh collection off to the db.

Max collections is set to the default of 3.

Takes:
	owner - string, the user that owns this
	name - string, the collections identifier in the user's space
*/

INSERT INTO users.collections 
(owner, name) 
VALUES
($1, $2)