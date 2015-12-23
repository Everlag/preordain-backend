/*

Returns all cards that have appeared in a given archetype
as long as that card was in a major event.

*/

select distinct(name) from mtgtop8.cards
	where deckid in
		(select deckid from mtgtop8.decks
			where name=$1 and
			eventid in
				(select eventid from major_events()))

	order by name;