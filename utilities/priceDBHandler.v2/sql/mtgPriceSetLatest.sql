/*
Returns the latest price for every card in a provided set.
*/

select * from prices.mtgprice where set=$1 and time=(select distinct(time) from prices.mtgprice where set=$1 order by time desc limit 1) order by price desc;