/*
	Watch files for changes, taken directly from gulp-starter
*/

var gulp   	= require('gulp');				// gulp runner

var config 	= require('../config.json');	// config json file	
var path   	= require('path');				// ability to join paths and the such
var watch  	= require('gulp-watch');		// do a set of tasks if something changes

var watchTask = function() {
  gulp.watch('src/**/*.ts', ['ts']); // watch for changes in typescript files
  gulp.watch('src/**/*.less', ['less']); // in less files
  gulp.watch(['src/**/*.html'], ['html']); // in every other file
}

gulp.task('watch', watchTask)
module.exports = watchTask