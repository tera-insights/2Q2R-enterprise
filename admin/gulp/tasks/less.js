var gulp 		= require('gulp'); 				// gulp task runner
var less 		= require('gulp-less'); 		// less compiler
var minifyCSS 	= require('gulp-minify-css'); 	// minify css before concat
var path 		= require("path"); 				// ability to merge paths and stuffs
var sourcemaps  = require('gulp-sourcemaps'); 	// sourcemaps to allow LESS debugging
var less 		= require('gulp-less'); 		// less compiler
var cleanCSS 	= require('gulp-clean-css');	// minify the css
var concat		= require('gulp-concat');		// concatenate all of our css files
var gutil		= require('gulp-util');			// used for logging, in this case error handling


var config      = require('../config.json');	// the config file

// paths of less files, grabbed from the config.json
var paths = {
	src: config.tasks.less.src,
	dest: config.tasks.less.dest
}

// the body of the less task
var lessTask = function () {
  return gulp.src(paths.src)
    .pipe(sourcemaps.init()) // start the sourcemap
    .pipe(less()) // less compiler
    .on('error', gutil.log) // if you get an error don't kill gulp-watch
    .pipe(cleanCSS()) // minify the css
    .pipe(concat('styles.min.css'))
    .pipe(sourcemaps.write()) // write the sourcemap
    .pipe(gulp.dest(paths.dest))
}

gulp.task('less', lessTask)
module.exports = lessTask
