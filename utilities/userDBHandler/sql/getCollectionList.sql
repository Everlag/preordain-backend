/*
Acquires the name of each collection a user has

Takes:
	owner - string, user that owns it
	collection - string, collection of that user
*/

SELECT
name, privacy
FROM
users.collections WHERE owner=$1