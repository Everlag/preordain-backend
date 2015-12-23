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

	eventid TEXT NOT NULL REFERENCES mtgtop8.events(eventid),

	/* There can only be one of a single deck! */
	CONSTRAINT uniqueDecksKey UNIQUE (deckid)
);

CREATE INDEX deck_name on mtgtop8.decks(name);
CREATE INDEX deck_deckId on mtgtop8.decks(deckid);
CREATE INDEX deck_eventid on mtgtop8.decks(eventid);


CREATE TABLE mtgtop8.cards (

	/* The full, normalized name of the card */
	name TEXT NOT NULL,

	quantity INT NOT NULL,

	/* If the card is present in the deck's sideboard */
	sideboard BOOLEAN NOT NULL,

	/* The unique mtgtop8 id for this deck */
	deckid TEXT NOT NULL REFERENCES mtgtop8.decks(deckid),

	/* A card can only appear once per deck in mainboard or sideboard */
	CONSTRAINT uniqueCardsKey UNIQUE (name, deckid, sideboard)

);

CREATE INDEX card_name on mtgtop8.cards(name);
CREATE INDEX card_deckId on mtgtop8.cards(deckid);

/*

Returns all eventids corresponding to
large events that represent the diverse metagame.

*/
CREATE FUNCTION major_events() RETURNS SETOF TEXT AS $$

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


drop function major_events();

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
