/*
Updates a collection on the database to have the desired permissions

Takes:
	name - string, user that owns it
	publicViewing - string, a valid privacy setting

*/

UPDATE users.collections
SET Privacy = $3
WHERE owner=$1 AND name=$2