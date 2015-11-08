/*
Acquires a subscription from the server with no authentication

Takes:
	name - string, user that owns it
*/

SELECT name, Plan, CustomerID, SubID, StartTime
FROM
users.subs WHERE name=$1