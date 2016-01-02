/*
To be run from psql. This sets up a complete postgres database for our prices.

NOTE: Comments must be block comments or they will break the run-ability of
      this setup.
	
*/

CREATE ROLE deckWriter WITH
	LOGIN
	ENCRYPTED
	PASSWORD '$insertPasswordHere';

CREATE ROLE deckReader WITH
	LOGIN
	ENCRYPTED
	PASSWORD '$insertPasswordHere';

/*
Create the database that we'll use.
*/
CREATE DATABASE deckData WITH
	OWNER     deckWriter 
	ENCODING 'UTF8';

COMMENT ON DATABASE deckData IS 'Where we keep our decks';


/*
Switch to the deckData database.

Execution may stop here; if it does, run the following command
and then bulk the remainder.

run 'psql deckData deckWriter' to connect as deckWriter
or
use '\connect deckData deckWriter' to connect as postgres
*/


/*
	Insertion order for a full set of event decks is
		.events
		.decks
		.cards
*/

/*
	Cleanup back to testing state
	DROP SCHEMA mtgtop8 CASCADE;
*/

/*Add a schema to work under*/
CREATE SCHEMA mtgtop8;


CREATE TABLE mtgtop8.events (

	/*
		The specific event's name
	*/
	name TEXT NOT NULL,
	
	eventid TEXT NOT NULL,

	happened TIMESTAMP NOT NULL,

	/* There can only be one of a single event! */
	CONSTRAINT uniqueEventsKey UNIQUE (eventid)
);

CREATE INDEX event_name on mtgtop8.events(name);
CREATE INDEX event_eventid on mtgtop8.events(eventid);
CREATE INDEX event_happened on mtgtop8.events(happened);


CREATE TABLE mtgtop8.decks (

	/*
		The archetype of the deck we determined
		based on metagame context, mtgtop8 rough archetype,
		and prescence of specific cards.
	*/
	name TEXT NOT NULL,
	player TEXT NOT NULL,

	deckid TEXT NOT NULL,

	parent TEXT NOT NULL REFERENCES mtgtop8.events(eventid),

	/* There can only be one of a single deck! */
	CONSTRAINT uniqueDecksKey UNIQUE (deckid)
);

CREATE INDEX deck_name on mtgtop8.decks(name);
CREATE INDEX deck_deckId on mtgtop8.decks(deckid);
CREATE INDEX deck_eventid on mtgtop8.decks(parent);


CREATE TABLE mtgtop8.cards (

	/* The full, normalized name of the card */
	name TEXT NOT NULL,

	quantity INT NOT NULL,

	/* If the card is present in the deck's sideboard */
	sideboard BOOLEAN NOT NULL,

	/* The unique mtgtop8 id for this deck */
	parent TEXT NOT NULL REFERENCES mtgtop8.decks(deckid),

	/* A card can only appear once per deck in mainboard or sideboard */
	CONSTRAINT uniqueCardsKey UNIQUE (name, parent, sideboard)

);

CREATE INDEX card_name on mtgtop8.cards(name);
CREATE INDEX card_deckId on mtgtop8.cards(parent);

/*

Returns all eventids corresponding to
large events that represent the diverse metagame.

*/
CREATE FUNCTION mtgtop8.major_events() RETURNS SETOF TEXT AS $$

	select eventid
	from mtgtop8.events
	where
		name like '%Grand Prix%' or
		name like '%Pro Tour%' or
		name like '%MKM Series%' or
		name like '%SCG%' or
		name like '%Modern MOCS%' or
		name like '%Modern Premier%'
	ORDER BY happened desc;

$$ LANGUAGE SQL IMMUTABLE;


/*

Returns all deckids corresponding to an archetype described by
an array of mtgtop8 names, cards used to exclude decks, and
cards required to be present in each decklist.

As a special case, a value of 'Default' in the cards to include
disables the requirement to have a card present.

This is a convenience function as copying this around for
individual queries is a special kind of hell to maintain.

*/
CREATE FUNCTION mtgtop8.archetype_decks(names TEXT[],
	badCards TEXT[], desiredCards TEXT[]) RETURNS SETOF TEXT AS $$

	select deckid from mtgtop8.decks
	/* Only major events */
	where parent in
		(select * from mtgtop8.major_events() as eventid) and
	/* Only in this archetype */
	name in
		(select unnest(names)) and
	/* Filter out decks with cards we don't want */
	not exists				
		(
			select * from
			(select unnest(badCards)) as has
			intersect
			(select name from mtgtop8.cards where parent=deckid)
		) and
	/* Filter out decks that fail to include cards we need */
	exists
		(
			select * from
			(select unnest(desiredCards)) as has
			intersect
			(select name from mtgtop8.cards where parent=deckid)
			/* Allow 'Default' input to ignore the above filter */
			union
			(select 'ignore' as default where 'Default' in
				(select * from unnest(desiredCards::text[]))
			)
		)


$$ LANGUAGE SQL IMMUTABLE;

/*
Privileges for the deckWriter
*/
REVOKE all privileges ON SCHEMA PUBLIC FROM deckWriter;
GRANT connect ON DATABASE deckData TO deckWriter;
GRANT usage ON SCHEMA PUBLIC TO deckWriter;
GRANT usage ON SCHEMA mtgtop8 TO deckWriter;

GRANT select, insert ON TABLE mtgtop8.cards to deckWriter;
GRANT select, insert ON TABLE mtgtop8.decks to deckWriter;
GRANT select, insert ON TABLE mtgtop8.events to deckWriter;

/*
Privileges for the deckReader
*/

REVOKE all privileges ON SCHEMA PUBLIC FROM deckReader;
GRANT connect ON DATABASE deckData TO deckReader;
GRANT usage ON SCHEMA PUBLIC TO deckReader;
GRANT usage ON SCHEMA mtgtop8 TO deckReader;

GRANT select ON TABLE mtgtop8.cards to deckReader;
GRANT select ON TABLE mtgtop8.decks to deckReader;
GRANT select ON TABLE mtgtop8.events to deckReader;
