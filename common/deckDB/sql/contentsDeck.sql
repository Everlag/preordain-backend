/*

Returns the contents of a deck given its deckid.

*/

select name, quantity::bigint, sideboard from mtgtop8.cards where parent=$1;