/*
Gets the highest price of the card across all printings for the most recent price of each printing.

See its lowest counterpart for an explanation of what magic it performs
*/



Select name, set, time, price, euro from 
(

SELECT DISTINCT ON(set) name, set, time, price, euro from prices.mtgprice
where name=$1 and price>0 and
set in
(

select distinct(set) from prices.mtgprice where name=$1

)
ORDER BY set, time DESC

) as temp ORDER BY price DESC LIMIT 1;