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

1. run `go-bindata -pkg="priceDB" sql` to regenerate bindings