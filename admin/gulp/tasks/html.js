/* 
	copies over html files 
*/
var gulp        = require('gulp');				// the task runner
var path        = require('path');				// path merging and associated functions
var changed		= require('gulp-changed');		// files that have been changed or not

var config 		= require('../config.json');	// configuration file

// paths used below, pulled from config.json
var paths = {
	src: config.tasks.html.src,
	dest: config.tasks.html.dest
}

// goes straight into the task
var htmlTask = function() {
  return gulp.src(paths.src)
    .pipe(changed(paths.src)) // ignore unchanged files
    .pipe(gulp.dest(paths.dest)); // and send them off
}

// just takes the task above instead of directly writing inside of the .task method
gulp.task('html', htmlTask);

module.exports = htmlTask
