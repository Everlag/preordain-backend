/*
Returns all prices for a printing of a specific card

Limits to 60 price points(~2 months) of data
*/

SELECT name, set, time, price FROM prices.mtgprice WHERE
name=$1 AND set=$2 order by time desc limit 60;