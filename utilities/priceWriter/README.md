Pricing is split into two sections.

priceWriter acquires price data from the various sources and writes that to
influxdb as a user with read/write permissions. Data from each source is a
tagged with the source name. priceWriter is run as a daemon

priceReader is embeddable and focuses on providing a readonly interface for
the influxdb. Queries are templated but readonly user permissions mitigates
all potential injection attacks.

Price data is organized in influxdb as follows
	Each card is represented by a time series. This represents all printings of the card
	Each time series has the columns {"time", "price", "set", "source"}
	time is the unix timestamp at second level resolution
	price is american cents
	set is the specific mtg expansion this price point has for this card
	source is the price data supplier this was acquired from

	
priceWriter:
	Setlist resides in setList.txt on disk in the directory it is in; the list
	is populated on import and is a newline delimited list of sets. Adding a
	set is as easy as adding a new line to the file and putting the set name in...
	assuming the appropriate vendor translation is in place for each source.
	It is the responsibility of the various price source components to translate
	the prices into their vendor specific representation - and there are a lot of them.
	
	The apiKeys are found in apiKeys.json and source code in priceWriter must be
	modified to add a new source. New sources files must also be added
	
	If the database is unavailable during a price data upload, that data is dumped to disk
	in the uploadFailures directory
	