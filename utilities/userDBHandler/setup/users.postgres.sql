/*
To be run from psql. This sets up a complete postgres database for our uses.

NOTE: Comments must be block comments or they will break the run-ability of
      this setup.


Windows: Start the environment with runsql.bat to get started

psql commands:
	\l - list databases
	\dt - list tables in current databases
	\du - list users
	\dD - list domains
	\df - list functions
	
To run freshly(as user postgres):
	'DROP DATABASE userdata;'
	'DROP USER usermanager;'
	
Login as usermanager with:
	'psql userdata usermanager'
*/

/*
Create the user that'll be our main mode of interaction.
Don't forget to set a reasonable password!
*/
CREATE ROLE usermanager WITH
	LOGIN
	ENCRYPTED
	PASSWORD '$insertPasswordHere';
	
/*
Create the database that we'll use.
*/
CREATE DATABASE userdata WITH
	OWNER     usermanager 
	ENCODING 'UTF8';

COMMENT ON DATABASE userdata IS 'Users, collections, and individual cards!';
	
/*
Switch to the userdata database.

Execution may stop here; if it does, run the following command
and then bulk the remainder.

run 'psql userdata usermanager' to connect as usermanager
or
use '\connect userdata postgres' to connect as postgres
*/

/*
All user submittable text is less than two tweets in length
*/
CREATE DOMAIN standardText TEXT CHECK (
    LENGTH(VALUE) < 280
);

/*
Qualities have very specific levels:
	Near Mint
	Lightly Played
	Heavily Played
*/
CREATE DOMAIN possibleQuality TEXT CHECK (
	VALUE = 'NM' OR
	VALUE = 'LP' OR
	VALUE = 'HP'
);

/*
Languages supported by the game as ISO 639-1 compliant
*/
CREATE DOMAIN possibleLanguage TEXT CHECK(
	VALUE = 'EN' OR /*English*/
	VALUE = 'ZH-HANS' OR /*Simplified Chinese*/
	VALUE = 'ZH-HANT' OR /*Traditional Chinese*/
	VALUE = 'FR' OR /*French*/
	VALUE = 'IT' OR /*Italian*/
	VALUE = 'DE' OR /*German*/
	VALUE = 'KO' OR /*Korean*/
	VALUE = 'JA' OR /*Japanese*/
	VALUE = 'PT' OR /*Portuguese*/
	VALUE = 'RU' OR /*Russian*/
	VALUE = 'ES' /*Spanish*/
);

/*
Privacy Settings we support for collections
*/
CREATE DOMAIN possiblePrivacy TEXT CHECK(
	VALUE = 'Private' OR
	VALUE = 'Contents' OR
	VALUE = 'History'
);

CREATE DOMAIN possibleSub TEXT CHECK(
	VALUE = 'Peek' OR
	VALUE = 'Preordain' OR
	VALUE = 'Sensei''s Top'
);

/*Add a schema to work under*/
CREATE SCHEMA users;

/*
Create the table holding lightweight user metadata.

A constraint is added to prevent multiple names from being the same.

passhash and nonce require a value.

No valid sessions or collections is the default state.

longestview is a duration, which is nanoseconds since epoch.
*/
CREATE TABLE users.meta (
	name standardText NOT NULL,
	email standardText NOT NULL,
	
	passhash bytea NOT NULL,
	nonce bytea NOT NULL,
	
	maxcollections int DEFAULT 1,
	longestview bigint DEFAULT 31560000000000000,
	
	CONSTRAINT uniquename UNIQUE (name)
);

CREATE UNIQUE INDEX meta_name_index on users.meta(name);
CREATE INDEX meta_email_index on users.meta(email);

/*
Create the table holding the user subscription information.

Adding a user should also fill in a new sub. A single sub entry
can exist per user, this avoids double-charging users.

Changing users.subs should, also, change the all applicable
fields for the user.
*/
CREATE TABLE users.subs (
	name standardText NOT NULL references users.meta(name),
	
	startTime timestamp NOT NULL,

	plan possibleSub NOT NULL,

	customerID TEXT NOT NULL,
	subID TEXT NOT NULL,
	
	CONSTRAINT unique_sub_name UNIQUE (name)
);

CREATE UNIQUE INDEX subs_name_index on users.subs(name);

/*
Create a function that allows us to mostly atomically upsert
into users.subs

select mod_sub('everlag', 'Sensei\'s Top', now,
	'someCustomerToken', 'someSubToken');
*/
CREATE FUNCTION
	mod_sub(specName TEXT, specPlan possibleSub, specTime timestamp,
			specCustomerID TEXT, specSubID TEXT)
	RETURNS VOID AS
$$
BEGIN
    LOOP
        -- first try to update the key
        UPDATE users.subs
			SET 
				startTime = specTime,
				plan = specPlan,
				customerID = specCustomerID,
				subID = specSubID
			WHERE
				name = specName;
        IF found THEN
            RETURN;
        END IF;
        -- not there, so try to insert the key
        -- if someone else inserts the same key concurrently,
        -- we could get a unique-key failure
        BEGIN
            INSERT INTO users.subs
				(name, startTime, plan, customerID, subID) 
			VALUES
				(specName, specTime, specPlan,
				specCustomerID, specSubID);
            RETURN;
        EXCEPTION WHEN unique_violation THEN
            -- do nothing, and loop to try the UPDATE again
        END;
    END LOOP;
END;
$$
LANGUAGE plpgsql;

/*
Create the table holding the user sessions.

The key is the primary way the session is accessed with the name
being included to require an associated identity.

Start must be less than end in order for this to be considered a valid
session. The following must be run regularly to prevent a buildup of invalid
sessions.

delete from usersessions where endValid <= startValid;
	
End should be updated rather than adding a new session.
*/
CREATE TABLE users.sessions (
	name standardText NOT NULL references users.meta(name),
	sessionKey bytea NOT NULL,
	
	startValid timestamp NOT NULL,
	endValid timestamp NOT NULL,
	
	CONSTRAINT uniqueSessionKey UNIQUE (sessionKey, name)
);

CREATE INDEX session_name_index on users.sessions(name);
CREATE INDEX session_key_index on users.sessions(sessionKey);

/*
Create our reset request. It is very similar to the sessions table
*/
CREATE TABLE users.resets (
	name standardText NOT NULL references users.meta(name),
	resetKey bytea NOT NULL,
	
	startValid timestamp DEFAULT now(),
	endValid timestamp DEFAULT (now()- INTERVAL '1 days'),
	
	CONSTRAINT uniqueResetKey UNIQUE (resetKey, name)
);


/*
Create the table that stores the collection metadata of our users.
*/
CREATE TABLE users.collections (

	name standardText NOT NULL,
	owner standardText NOT NULL references users.meta(name),
	
	lastUpdate timestamp DEFAULT now(),
	
	Privacy possiblePrivacy DEFAULT 'Contents',

	CONSTRAINT uniqueCollectionKey UNIQUE (name, owner)
);

/*
A table that stores the actual contents of the collection.

These contents are the most up to date.

Uniquely index by ownere:collection:cardName:setName:quality

Update using UPSERT... which needs to be implemented
*/
CREATE TABLE users.collectionContents (

	cardName standardText NOT NULL,
	setName standardText NOT NULL,
	comment standardText NOT NULL,
	
	
	quantity int NOT NULL,
	quality possibleQuality NOT NULL,
	lang possibleLanguage NOT NULL,
	
	owner standardText NOT NULL,
	collection standardText NOT NULL,
	
	lastUpdate timestamp NOT NULL,

	FOREIGN KEY (owner, collection) REFERENCES users.collections (owner, name),

	CONSTRAINT uniqueContentsKey UNIQUE (owner, collection,
										cardName, setName,
										quality, lang)
);

CREATE INDEX contents_completeCollection_index on users.collectionContents(owner, collection);

/*
A table that stores the changes each collection undergoes.

This single table allows us to rebuild the complete, publicly visible
database.

NOTICE: No foreign key dependency as we want to be capable of rebuilding from
        this single table.

        Thus, we track when a row was created to avoid potential future issues
*/
CREATE TABLE users.collectionHistory (

	cardName standardText NOT NULL,
	setName standardText NOT NULL,
	comment standardText NOT NULL,
	
	quantity int NOT NULL,
	quality possibleQuality NOT NULL,
	lang possibleLanguage NOT NULL,
	
	owner standardText NOT NULL,
	collection standardText NOT NULL,
	
	lastUpdate timestamp NOT NULL,

	creationTime timestamp DEFAULT now(),

	CONSTRAINT uniqueHistoryKey UNIQUE (owner, collection,
										cardName, setName,
										quality, lang,
										lastUpdate)
);

CREATE INDEX history_cardName_index on users.collectionHistory(cardName, owner);

/*
Create a function that allows us to mostly atomically upsert
into userCollectionContents

select add_card('bleh', 'bleh', 'bleh', 'bleh', 'bleh', 3, 'NM');
drop function add_card(TEXT, TEXT, TEXT, TEXT, TEXT, INT, possibleQuality);
*/
CREATE FUNCTION
	add_card(specOwner TEXT, specCollection TEXT,
			specCardName TEXT, specSetName TEXT, specComment TEXT,
			specQuantity int,
			specQuality possibleQuality, specLang possibleLanguage,
			specTime timestamp)
	RETURNS VOID AS
$$
BEGIN
    LOOP
        -- first try to update the key
        UPDATE users.collectionContents
			SET 
				comment = specComment,
				quantity = quantity + specQuantity,
				lastUpdate = specTime
			WHERE
				cardName = specCardName AND
				setName = specSetName AND
				quality = specQuality AND
				lang = specLang AND
				owner = specOwner AND
				collection = specCollection;
        IF found THEN
            RETURN;
        END IF;
        -- not there, so try to insert the key
        -- if someone else inserts the same key concurrently,
        -- we could get a unique-key failure
        BEGIN
            INSERT INTO users.collectionContents
				(owner, collection, cardName, setName, comment,
					quantity, quality, lang, lastUpdate) 
			VALUES
				(specOwner, specCollection, specCardName,
				specSetName, specComment,
				specQuantity, specQuality, specLang, specTime);
            RETURN;
        EXCEPTION WHEN unique_violation THEN
            -- do nothing, and loop to try the UPDATE again
        END;
    END LOOP;
END;
$$
LANGUAGE plpgsql;


/*
Lock all permissions down to minimum.

NOTE: Do this as user 'postgres' in the userdata table!

*Assume select for each*
users.meta - insert and update
users.Sessions - insert and delete
users.Resets - insert and delete
users.Collections - insert, update, and delete
users.CollectionContents - insert and update
users.CollectionHistory - insert
*/

/*Make sure all permissions are OFF by default*/
REVOKE all privileges ON SCHEMA PUBLIC FROM usermanager;
REVOKE all ON DATABASE POSTGRES FROM usermanager;
REVOKE all ON DATABASE userdata FROM usermanager;
REVOKE create ON DATABASE userdata FROM usermanager;

/*Need to use to do anything else*/
GRANT connect ON DATABASE userdata TO usermanager;
GRANT usage ON SCHEMA PUBLIC TO usermanager;
GRANT usage ON SCHEMA users TO usermanager;

/*Whitelist all usage per table*/
GRANT select, insert, update ON TABLE users.meta to userManager;

/*Subs cannot be deleted, only added or altered*/
GRANT select, insert, update ON TABLE users.subs to userManager;

/*Sessions and resets can be deleted with no issue*/
GRANT select, insert, delete ON TABLE users.sessions to userManager;
GRANT select, insert, delete ON TABLE users.resets to userManager;

/*Collections needs to be capable of being deleted*/
GRANT select, insert, update, delete ON TABLE users.collections to userManager;

GRANT select, insert, update ON TABLE users.collectionContents to userManager;

/*Append only collection history is VERY important*/
GRANT select, insert ON TABLE users.collectionHistory to userManager;

/*
Set a backup user up so we are able to remotely dump table contents and
nothing else.

Backups can be safely run using
'pg_dump -d userdata -U usermanager -f TARGET.sql'
or -f omitted and piped to gzip!
*/

CREATE ROLE backupper WITH
	LOGIN
	ENCRYPTED
	PASSWORD '$insertAnotherPasswordHere';

GRANT usage ON SCHEMA public, users to backupper; 
GRANT connect ON DATABASE userdata TO backupper;
GRANT select ON ALL TABLES IN SCHEMA users TO backupper;