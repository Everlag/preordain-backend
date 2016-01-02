# priceWriter

deckWriter acquires new modern event data from mtgtop8 daily.

## State Notice

Some state is stored in a file that sits inside the working directory of the process. This allows us to avoid any potentially abusive scraping.

As this state is neither sensitive nor overly important, it can be treated much like log files are.

## Environment Notice

Two environment variables must be present when calling this package.

1. `POSTGRES_CONFIG` — location of package config

1. `POSTGRES_CERT` — location of postgres cert to trust
