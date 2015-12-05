# preordain-backend

This powers the backend over at [preorda.in](https://preorda.in).

## Organization

This repo is organized into several distinct areas.

+ api - RESTful api designed to be daemonized. Each directory is a specific api. 

+ utilities - a series of utilities that handle developing and acquiring static content in the site.

+ common - small packages used to centralize access around specific resources.


Dependencies between major sections are avoided. Additionally, avoid co-dependencies between subsections of major branches; move them to common.

## Deployment

All packages have a number of environment variables which allow consolidating common resource dependencies. These include `SETLIST`, `MTGJSON`, and others.

Each package is standardized around assuming that any unset environment variable means the current working directory. This is useful for development but obviously goes against best practices in production.