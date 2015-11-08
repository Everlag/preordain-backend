/*
Returns all prices for a printing of a specific card
*/

SELECT name, set, time, price FROM prices.mtgprice WHERE
name=$1 AND set=$2