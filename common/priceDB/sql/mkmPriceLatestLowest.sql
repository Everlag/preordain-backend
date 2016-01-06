/*
Get the lowest, positive price of a card across all of its printings's
latest prices.
*/

with tiny_window as (
	select * from prices.magiccardmarket
	where
		name=$1 and
		now() - time < '1 week'::interval and
		price > 0
),
all_sets as (
	select set from tiny_window where name=$1 group by set
)
select name, set, time, price from(

	SELECT DISTINCT ON(set) name, set, time, price from tiny_window
	where
		set in (select * from all_sets)
	order by set, time desc

) as temp order by price desc limit 1;