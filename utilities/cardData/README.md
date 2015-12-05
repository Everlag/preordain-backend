# cardData

cardData builds content related to Magic cards.

The architecture is modular. Adding another relative package, necessary datums to store in cardData/card(), and a line in cardData/getAllCardData() will allow integration of a new section of data. Packages which don't simply plug in can require a slightly more in depth integration, call them in cardInfo/main() after everything else.

## Deployment notice

Any directories which are populated, such as cardData or typeAhead, must exist prior to starting the program. This includes cache locations as well as output.

Special attention should be paid to ensuring cache files, *.cache.*, remain across program runs unless you want a lengthy scrape of mtgsalvation and mtgtop8. Prompting a cache refresh can be done by removing the cache file.

## Environment Notice

Several optional environment variables are provided for configuration

1. `MTGJSON` — location of mtgjson generated card data

1. `SETLIST` — location of set list to use.

1. `CACHE`   — location of intermediate cache files. 

1. `OUTPUT`   — location of output directory. 

These remaining unset will result in all actions happening relative to the CWD of that process.