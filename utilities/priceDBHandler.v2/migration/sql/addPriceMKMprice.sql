/*

Adds a new price point for a specific card and printing of that
card to prices.magiccardmarket

*/

INSERT INTO prices.magiccardmarket
(name, set, time, euro, price)
values
($1, $2, $3, $4, $5)