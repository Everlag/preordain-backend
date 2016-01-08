/*
To be run from psql. This sets up a complete postgres database for our prices.

NOTE: Comments must be block comments or they will break the run-ability of
      this setup.

CREATE ROLE usermanager WITH
	LOGIN
	ENCRYPTED
	PASSWORD '$insertPasswordHere';
	
*/

CREATE ROLE priceWriter WITH
	LOGIN
	ENCRYPTED
	PASSWORD '$insertPasswordHere';

CREATE ROLE priceReader WITH
	LOGIN
	ENCRYPTED
	PASSWORD '$insertPasswordHere';

/*
Create the database that we'll use.
*/
CREATE DATABASE priceData WITH
	OWNER     priceWriter 
	ENCODING 'UTF8';

COMMENT ON DATABASE priceData IS 'Where we keep our prices';


/*
Switch to the priceData database.

Execution may stop here; if it does, run the following command
and then bulk the remainder.

run 'psql priceData priceWriter' to connect as priceWriter
or
use '\connect pricedata priceWriter' to connect as postgres
*/

/*Add a schema to work under*/
CREATE SCHEMA prices;


/*

Add the tables that shall store our prices.

*/
CREATE TABLE prices.mtgprice (

	name TEXT NOT NULL,
	set TEXT NOT NULL,
	time timestamp NOT NULL,

	price int NOT NULL,

	CONSTRAINT uniqueMTGpriceEntryKey UNIQUE (name, set, time)

);

CREATE INDEX mtgprice_name_index on prices.mtgprice(name);
CREATE INDEX mtgprice_set_index on prices.mtgprice("set");
CREATE INDEX mtgprice_time_index on prices.mtgprice("time");

CREATE TABLE prices.magiccardmarket (

	name TEXT NOT NULL,
	set TEXT NOT NULL,
	time timestamp NOT NULL,

	price int NOT NULL,
	euro int NOT NULL,

	CONSTRAINT uniqueMKMEntryKey UNIQUE (name, set, time)

);

CREATE INDEX mkm_name_index on prices.magiccardmarket(name);
CREATE INDEX mkm_set_index on prices.magiccardmarket("set");
CREATE INDEX mkm_time_index on prices.magiccardmarket("time");

/*The aggregate median is very necessary*/
CREATE FUNCTION _final_median(anyarray) RETURNS int AS $$ 
  WITH q AS
  (
     SELECT val
     FROM unnest($1) val
     WHERE VAL IS NOT NULL
     ORDER BY 1
  ),
  cnt AS
  (
    SELECT COUNT(*) AS c FROM q
  )
  SELECT AVG(val)::int
  FROM 
  (
    SELECT val FROM q
    LIMIT  2 - MOD((SELECT c FROM cnt), 2)
    OFFSET GREATEST(CEIL((SELECT c FROM cnt) / 2.0) - 1,0)  
  ) q2;
$$ LANGUAGE SQL IMMUTABLE;
 
CREATE AGGREGATE median(anyelement) (
  SFUNC=array_append,
  STYPE=anyarray,
  FINALFUNC=_final_median,
  INITCOND='{}'
);


/*
	Create a function and associated type that allows us to hit
	bulk queries in a single db trip for latest extrema.

	This is not overly efficient but keeps the work in the database
	as a single client-facing query rather than len(cards)
	queries.
*/
create type latestitem as (name text, set text, time timestamp, price int);

CREATE OR REPLACE FUNCTION forEachLatest(IN prepared text, IN cards text[])
  RETURNS setof latestitem AS

  $$
  DECLARE
    card text;

    query text;

    latest latestitem%rowtype;
  BEGIN

    foreach card in array cards
    loop

      query = format('execute %s(%s)',
      	quote_ident(prepared), quote_literal(card));
      execute query into latest;

      return next latest;

    end loop;
  
    RETURN;
  END;
  $$

LANGUAGE plpgsql VOLATILE;


/*
Privileges for the pricewriter
*/
REVOKE all privileges ON SCHEMA PUBLIC FROM priceWriter;
GRANT connect ON DATABASE priceData TO priceWriter;
GRANT usage ON SCHEMA PUBLIC TO priceWriter;
GRANT usage ON SCHEMA prices TO priceWriter;

GRANT select, insert ON TABLE prices.magiccardmarket to priceWriter;
GRANT select, insert ON TABLE prices.mtgprice to priceWriter;

/*
Privileges for the pricereader
*/

REVOKE all privileges ON SCHEMA PUBLIC FROM priceReader;
GRANT connect ON DATABASE priceData TO priceReader;
GRANT usage ON SCHEMA PUBLIC TO priceReader;
GRANT usage ON SCHEMA prices TO priceReader;

GRANT select ON TABLE prices.magiccardmarket to priceReader;
GRANT select ON TABLE prices.mtgprice to priceReader;