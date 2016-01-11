/*
Returns the weekly median of extreme prices summed for a collection of cards.

This query is designed for use with statements following the weekly-extrema
pattern of return values.

Takes
	$1 - prepared statement identifier
	$2 - array of card names to fetch
	$3 - array of integers to multiply the corresponding card's price by

Constraint
	len($2) == len($3) or things start breaking.

The fact that we prepare all statements before we use a new connection
means this query can be safely run, anything that invalidates that assumption
will break this query.

Here's a strange scapeshift build with the assumption that a weekslow
statement has been prepared.

select * from summedWeeklyExtrema('weeklow',
array['Breeding Pool', 'Cryptic Command', 'Farseek', 'Flooded Grove', 'Forest', 'Island', 'Mountain', 'Peer Through Depths', 'Pyroclasm', 'Remand', 'Repeal', 'Sakura-Tribe Elder', 'Scapeshift', 'Search for Tomorrow', 'Serum Visions', 'Snapcaster Mage', 'Steam Vents', 'Stomping Ground', 'Valakut, the Molten Pinnacle', 'Ancient Grudge', 'Firespout', 'Negate', 'Rude Awakening', 'Shadow of Doubt', 'Sudden Shock', 'Wurmcoil Engine']::text[],
array['1', '3', '2', '1', '4', '3', '3', '3', '3', '4', '2', '4', '4', '4', '4', '3', '4', '4', '4', '3', '2', '3', '1', '2', '3', '1']::int[]
);

*/

select week, price from summedWeeklyExtrema($1, $2, $3);