/*
Acquires the every valid reset key for a provided user.

Takes:
	name - string, user that owns it
*/

SELECT name, resetKey, startValid, endValid
FROM users.resets
WHERE name=$1 AND endValid > now()