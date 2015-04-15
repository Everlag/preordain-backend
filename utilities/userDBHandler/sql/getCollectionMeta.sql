/*
Acquires the metadata of a user's collection.

Takes:
	owner - string, user that owns it
	collection - string, collection of that user
*/

SELECT
name, owner, lastUpdate, publicViewing, publicHistory
FROM
users.collections WHERE owner=$1 AND name=$2