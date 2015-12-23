/*

Returns all deck archetypes a card was included in
for major events.

*/

select name from mtgtop8.decks where
	deckid in 
		(select deckid from mtgtop8.cards where name=$1)
	and eventid in
		(select * from major_events())
	group by name;