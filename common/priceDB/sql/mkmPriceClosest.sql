/*
Returns the price for a given card/set combination
with time closest to the provided time.

The acquired price will ALWAYS be after or at the time provided. That means a query before the start of data will return nothing. This is the behaviour we desire as the alternative would mean assigning a single price for all dates for data start for a card/set.
*/

select name, set, time, price, euro from prices.magiccardmarket
where name=$1 and set=$2 and $3 > time
order by
abs(extract(epoch from $3 - time))
limit 1;