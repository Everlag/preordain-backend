/*

Adds a new deck with all of its necessary metadata to
mtgtop8.decks

*/

INSERT INTO mtgtop8.decks
(name, player, deckid, parent)
values
($1, $2, $3, $4)