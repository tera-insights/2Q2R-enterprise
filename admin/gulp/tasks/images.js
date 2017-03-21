var gulp        = require('gulp');				// the task runner
var imagemin    = require('gulp-imagemin');		// minify images to reduce load
var path        = require('path');				// path merging and associated functions
var changed		= require('gulp-changed');		// files that have been changed or not

var config 		= require('../config.json');	// configuration file

// paths used below, pulled from config.json
var paths = {
	src: config.tasks.images.src,
	dest: config.tasks.images.dest
}

// goes straight into the task
var imagesTask = function() {
  return gulp.src([paths.src, '*!README.md'])
    .pipe(changed(paths.dest)) // Ignore unchanged files
    .pipe(imagemin()) // Optimize
    .pipe(gulp.dest(paths.dest))
}

// just takes the task above instead of directly writing inside of the .task method
gulp.task('images', imagesTask);

module.exports = imagesTask
