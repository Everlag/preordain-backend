# cardCropper

Crops a provided directory of full cards to just the art.

## Usage

Directory is fed in with --src, you'll get yelled at if you don't.

Additionally, height and width each must be provided with --height and --width.

ie

	`gulp --src ./testingGround --height 680 --width 480`

Results are placed inside ./dist.

## Content Notice

This provides a 'best effort' approach to cropping, this is not a replacement for the 300dpi crops made available by CCGHQ. Typical use case is cropping the low quality card art provided before set release while waiting for full quality crops to be released.

## Sources

Uses [gulp-jimp](https://github.com/haydenbleasel/gulp-jimp) for an easy to deploy method of manipulating images in node. This is slower than other solutions but has no large, binary dependencies. 