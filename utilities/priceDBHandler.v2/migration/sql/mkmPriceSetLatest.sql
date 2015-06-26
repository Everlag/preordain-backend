/*
Returns the latest price for every card in a provided set.
*/

SELECT name, set, time, price, euro from(SELECT DISTINCT ON(name) name, set, time, price, euro from prices.magiccardmarket where set=$1 order by name, time DESC) as temp;
