/*
Returns the weekly aggregate median for a name and set combination
*/
select date_trunc('week', time), median(price) from prices.magiccardmarket
where
name=$1 and set=$2
group by date_trunc('week', time) order by date_trunc('week', time) desc;