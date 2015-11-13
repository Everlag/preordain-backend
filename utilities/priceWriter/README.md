# priceWriter

priceWriter acquires price data from mtgprice and magiccardmarket every day.

## Environment Notice

Two environment variables must be present when calling this package.

1. `POSTGRES_CONFIG` — location of package config

1. `POSTGRES_CERT` — location of postgres cert to trust

Additionally, two optional environment variables are provided for configuration

1. `APIKEYS` — location of apiKeys.json

1. `SETLIST` — location of set list to use.