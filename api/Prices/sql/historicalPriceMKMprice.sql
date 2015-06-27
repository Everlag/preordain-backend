/*
Returns all prices for a printing of a specific card
*/

SELECT name, set, time, price, euro FROM prices.magiccardmarket WHERE
name=$1 AND set=$2