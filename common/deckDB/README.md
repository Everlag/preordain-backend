# Structure portability

The structures found in this package are the authoritative structures supported by the remote db.

Scrapers and such are recommended to import deckDB and use these structures to avoid having to translate their local structs into these normalized versions.

# Format Support

Currently we handle only a single format, modern. It would be relatively painless to extend this to handle more than modern but the stability of the format as well as medium entry cost makes it a good place to start.

# Event Classification

While the database stores every event, the queries this package expose will only ever present major events.

Major events are defined as containing one of these substrings.

+ Grand Prix

+ Pro Tour

+ MKM Series

+ SCG

+ Modern Premier

+ Modern MOCS

# Environment Notice

Two environment variables must be present when calling this package.

1. `POSTGRES_CONFIG` specifies location of package config

1. `POSTGRES_CERT`specifies location of postgres cert to trust

# Deployment Notes
	
Copy sql into directory beside binary. The sql present in this package's sql subdirectory is the authoritative version

Create certs directory beside binary and follow instructions in testing for generating the trust chain.

**This package is not safe for use with anything except self signed certificates where the root ca is equal to the server certificate.**
	
# Development Notes

go-bindata is used to avoid having to copy the sql to every user of the data.

Ensure go-bindata is installed: `go get -u github.com/jteeuwen/go-bindata/...`

When adding or editing sql:
1. put inside sql directory as 'handle'.sql

1. it must be added to the dbHandler as a constant then added to the statements.

1. run `go-bindata -pkg="deckDB" sql` to regenerate bindings