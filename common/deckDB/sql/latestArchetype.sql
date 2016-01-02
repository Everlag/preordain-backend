/*

Returns the deckid of the latest appearance of the archetype at a major event
alongside relevant data describing the event and deck.


Takes
	$1 array of normalized mtgtop8 archetype names
	$2 array of card names to explicitly require not be present in decks
	$3 array of card names to explicitly require in decks, if 'Default'
	   is present inside the array then the filter is

*/

select mtgtop8.decks.deckid, mtgtop8.decks.player,
	mtgtop8.events.name, mtgtop8.events.happened

from mtgtop8.decks, mtgtop8.events
where deckid in
	
	(select deckid from mtgtop8.archetype_decks($1, $2, $3) as deckid) and

	mtgtop8.decks.parent = mtgtop8.events.eventid

order by mtgtop8.events.happened desc limit 1;