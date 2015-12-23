/*

Adds a new event with all of its necessary metadata to
mtgtop8.events

*/

INSERT INTO mtgtop8.events
(name, eventid, happened)
values
($1, $2, $3)