# preorda.in Deck API

The deck api exposes a number of endpoints for acquiring information about magic decks that have placed in major tournaments.

## Generated docs

Documentation is exposed as a swagger config on api/Decks/apidocs.json

This documentation can be loaded into any swagger 1.2 compatible viewer.

## Environment Notice

To environment variables must be present when calling this package.

1. `POSTGRES_CONFIG` — location of package config

1. `POSTGRES_CERT` — location of postgres cert to trust

All environment variables have sane defaults for *development*. These defaults are provided in `prices.default.env`. They should be explicitly specified when operation in production.

## Authentication and abuse

This version of the api, 0.2, is unauthenticated.

Note that ip and referer are checked against againt abuse in middleware. Excessive api use without being whitelisted will result in throttling and, eventually, an ip ban. To be whitelisted, contact this repo's owner.