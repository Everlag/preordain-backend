/*

Adds a new card with all of its necessary metadata to
mtgtop8.cards

*/

INSERT INTO mtgtop8.cards
(name, quantity, sideboard, parent)
values
($1, $2, $3, $4)