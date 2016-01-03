/*

Returns the metadata of a deck given its deckid

*/

select player, name from mtgtop8.decks where deckid=$1;