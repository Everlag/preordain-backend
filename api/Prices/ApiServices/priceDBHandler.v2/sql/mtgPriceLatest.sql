/*
Returns the latest update for a card/set combination.
*/

SELECT name, set, time, price FROM prices.mtgprice
WHERE name=$1 AND set=$2
ORDER BY time DESC LIMIT 1;