/*
Gets the lowest(non-zero) price of the card across all printings for the most recent price of each printing.

This is a complex query that does a lot of work for us right in the database,
I break it down here

get the most recent price of each set of the card, distinct on allows this
eachSetLatest =
	SELECT DISTINCT ON(set) name, set, time, price from prices.magiccardmarket where name=$1 and set in theSets

get each set of the card we have on record
theSets =
	(select distinct(set) from prices.magiccardmarket where name=$1)

Select name, set, time, price from (eachSetLatest) ORDER BY set, time DESC) as temp;

*/

Select name, set, time, price, euro from 
(

SELECT DISTINCT ON(set) name, set, time, price, euro from prices.magiccardmarket
where name=$1 and price>0 and
set in
(

select distinct(set) from prices.magiccardmarket where name=$1

)
ORDER BY set, time DESC

) as temp ORDER BY price ASC LIMIT 1;