/*

Adds a new price point for a specific card and printing of that
card to prices.mtgprice

*/

INSERT INTO prices.mtgprice
(name, set, time, price)
values
($1, $2, $3, $4)