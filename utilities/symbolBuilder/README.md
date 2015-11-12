# symbolBuilder

Minifies a given directory of svgs.

## Usage

Directory is fed in with --src, you'll get yelled at if you don't.

ie

	gulp --src ../cardSymbols

Results are placed inside ./dist and ./dist/dist.zip is made for deployment convenience.

Expect the total size, which is reported pre- and post-minify, to be nearly half of the unoptimized versions.

## Content Notice

The cardSymbols directory provided with this package is the authoritative collection of svg MTG card symbols used in preorda.in

If a symbol is added to the project, it is to be added here.

## Sources

Uses [gulp-svgmin](https://github.com/ben-eb/gulp-svgmin) which provides easy access to [svgo](https://github.com/svg/svgo). All plugins are set to their defaults as they do a perfectly fine job!