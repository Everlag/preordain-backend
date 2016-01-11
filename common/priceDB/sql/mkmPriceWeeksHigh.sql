/*
Get the lowest, positive price of a card for every week across
all of it's printings.
*/

with weekly as (
	select
		date_trunc('week', time) as week,
		median(price) as median,
		set
	from prices.magiccardmarket
	where
		name=$1 and
		price > 0
	group by
		date_trunc('week', time), set
	order by
		date_trunc('week', time) desc
),
weekly_min as (
	select week, max(median) from weekly group by week
)
select $1 as name, * from weekly_min order by week desc;