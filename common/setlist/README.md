# setlist

This package allows for easy use of set lists in the format defined by this package.

Set lists have the following properties

1. Named setList.txt

1. Sets are presented as a newline delimited list

1. Blank lines are allowed.

1. Foil variants of sets are defined as having a ` Foil` suffix.

1. Sets follow the naming in [mtgjson](http://mtgjson.com/). If a set is misspelled in mtgjson, we follow that spelling.

## Environment config

The package defaults to fetching a set list from the current working directory.

The `SETLIST` environment variable can optionally specify which directory the list can be found in. This is useful for centralizing configuration across various services.