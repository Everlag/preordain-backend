/*
Returns all prices for a printing of a specific card
*/

SELECT * FROM prices.mtgprice WHERE
name=$1 AND set=$2