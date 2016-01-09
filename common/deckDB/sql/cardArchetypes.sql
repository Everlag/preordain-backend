/*

Returns all deck archetypes a card was included in
for major events.

*/

with deckids as (
	select parent from mtgtop8.cards
	where
		name = $1
	group by parent
),
majors as (
	select * from mtgtop8.major_events()
)

select name from mtgtop8.decks
	where
		deckid in (select * from deckids)
	and 
		parent in (select * from majors)
group by name;