'use strict';

// Include Gulp & Tools We'll Use
let gulp = require('gulp');

let del = require('del');

let jimp = require('gulp-jimp');

let rename = require('gulp-rename');

let argv = require('yargs').argv;

// Clean Output Directory
gulp.task('clean', del.bind(null, 'dist'));

// Crop images from cards, the Default Task
//
// Assumes post-M15 card frame
gulp.task('default', ['clean'], function () {
  let src = argv.src;
  if (src === undefined) throw 'src directory must be defined with --src';

  // We shamefully ask the user for image dimensions,
  // jimp doesn't support % based crops.
  let inHeight = argv.height;
  if (inHeight === undefined) throw 'image height must be explicitly provided';

  let inWidth = argv.width;
  if (inWidth === undefined) throw 'image width must be explicitly provided';

  // Pixel measurements taken from post-M15 frame 
  let refWidth = 312;
  let refHeight = 445;
  let x = 20; // Start of crop on each axis
  let y = 48;
  let width = 273; // Length of crop on each axis
  let height = 198;

  // Convert to input dimensions by using a ratio
  let widthRatio = inWidth / refWidth;
  let heightRatio = inHeight / refHeight;
  x*= widthRatio;
  y*= heightRatio;
  width*= widthRatio;
  height*= heightRatio;

  return gulp.src(`${src}/*.full.jpg`)
    // Crop what should be the art
  	.pipe(jimp({
        '': {
            crop: { x: x, y: y, width: width, height: height },
        },
    }))
    // Rename from common full card naming
    // to typical crop format
    .pipe(rename(function (path) {
      path.basename = path.basename
        .replace('.full', '')
        .toLowerCase();
      return path;
    }))
    .pipe(gulp.dest('./dist'));

});