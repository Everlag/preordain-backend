# preorda.in Price API

The price api exposes a series of endpoints that cover the majority of what a magic site could desire.

## Generated docs

Documentation is exposed as a swagger config on api/Prices/apidocs.json

This documentation can be loaded into any swagger 1.2 compatible viewer.

## Environment Notice

Three environment variables must be present when calling this package.

1. `POSTGRES_CONFIG` — location of package config

1. `POSTGRES_CERT` — location of postgres cert to trust

1. `DECK_API` — local port the remote deck api sits on

Additionally, two optional environment variables are provided for configuration

1. `MTGJSON` — location of mtgjson generated card data

1. `SETLIST` — location of set list to use.

All environment variables have sane defaults for *development*. These defaults are provided in `prices.default.env`. They should be explicitly specified when operation in production.

## Authentication and abuse

This version of the api, 0.2, is unauthenticated.

Note that ip and referer are checked against againt abuse in middleware. Excessive api use without being whitelisted will result in throttling and, eventually, an ip ban. To be whitelisted, contact this repo's owner.