/*
Sends a verified user off to the database.

Max collections is set to the default of 3.

Takes:
	name - string, user that owns it
	email - string, the address we can contact a user at
	passHash - bytea, the result of scrypt(password, nonce)
	nonce - bytea, the nonce used for deriving passHash
*/

INSERT INTO users.meta 
(name, email, passHash, nonce) 
VALUES
($1, $2, $3, $4)