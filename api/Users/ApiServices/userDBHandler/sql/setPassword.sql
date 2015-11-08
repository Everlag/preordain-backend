/*
Updates a user's password on the database

Takes:
	name - string, user that owns it
	passHash - bytea, the result of scrypt(password, nonce)
	nonce - bytea, the nonce used for deriving passHash
*/

UPDATE users.meta
SET passHash = $2, nonce=$3
WHERE
name=$1