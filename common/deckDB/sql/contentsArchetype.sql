/*

Returns all cards that have appeared in a given archetype
as long as that card was in a major event. This also returns
how many times each card appeared, a useful indicator of use.


Takes
	$1 array of normalized mtgtop8 archetype names
	$2 array of card names to explicitly require not be present in decks
	$3 array of card names to explicitly require in decks, if 'Default'
	   is present inside the array then the filter is ignored.
*/


select name, sum(quantity) from mtgtop8.cards
	where parent in
		(select * from mtgtop8.archetype_decks($1, $2, $3))
	group by name order by sum(quantity) desc;