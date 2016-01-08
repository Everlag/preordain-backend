/*
Returns a the result of querying a prepared statement for each card
provided.

This query is designed for use with statements following the latest-extrema
pattern of return values.

Takes
	$1 - prepared statement identifier
	$2 - array of card names to fetch

The fact that we prepare all statements before we use a new connection
means this query can be safely run, anything that invalidates that assumption
will break this query.

Here's a subset of Jund for a usage example, with mtgPriceLatestLowest
assumed to be prepared for the session.

select name, set,  price from forEachLatest('mtgPriceLatestLowest',
	array['Lightning Bolt', 'Overgrown Tomb', 'Stomping Ground', 'Forest',
	'Swamp',  'Wooded Foothills', 'Bloodstained Mire', 'Blackcleave Cliffs',
	'Raging Ravine', 'Verdant Catacombs', 'Kitchen Finks',
	'Tasigur, the Golden Fang', 'Scavenging Ooze', 'Dark Confidant',
	'Tarmogoyf', 'Liliana of the Veil', 'Kolaghan''s Command',
	'Abrupt Decay']::text[]);

*/

select name, set, time, price from forEachLatest($1, $2);