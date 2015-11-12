'use strict';

// Include Gulp & Tools We'll Use
let gulp = require('gulp');

let del = require('del');

let svgmin = require('gulp-svgmin');
let size = require('gulp-size');

let argv = require('yargs').argv;

let zip = require('gulp-zip');

// Clean Output Directory
gulp.task('clean', del.bind(null, 'dist'));

// Build Production Files, the Default Task
gulp.task('default', ['clean'], function () {
  let src = argv.src;
  if (src === undefined) throw 'src directory must be defined with --src';

  return gulp.src(`${src}/*.svg`)
    .pipe(size())
    .pipe(svgmin())
    .pipe(size())
    .pipe(gulp.dest('./dist'))
    .pipe(zip("dist.zip")) // To the archive
    .pipe(gulp.dest('./dist'));

});