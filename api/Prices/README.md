# preorda.in Price API

The price api exposes a series of endpoints that cover the majority of what a magic site could desire.

## Generated docs

Documentation is exposed as a swagger config on api/Prices/apidocs.json

This documentation can be loaded into any swagger 1.2 compatible viewer.

## Authentication and abuse

This version of the api, 0.2, is unauthenticated.

Note that ip and referer are checked against againt abuse in middleware. Excessive api use without being whitelisted will result in throttling and, eventually, an ip ban. To be whitelisted, contact this repo's owner.