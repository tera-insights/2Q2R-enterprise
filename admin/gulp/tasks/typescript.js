/*
	For compilation and sourcemapping of typescript files
*/

var gulp 			= require('gulp');					// the task runner
var sourcemaps 		= require('gulp-sourcemaps');		// sourcemaps to make debugging sane
var browserify 		= require('browserify');			// module system
var ts 				= require('gulp-typescript');		// bring in config file
var tsify 			= require('tsify');					// browserify typescript compilation
var source 			= require('vinyl-source-stream');	// needed for what we do with browserify
var uglify 			= require('gulp-uglify');			// minimize bundled script size
var buffer 			= require('vinyl-buffer');			// needed to do most operations on a vinyl stream

var config 			= require('../config.json');		// config file that controls src/dest (mostly)

// This is the config file for typescript compilation
// DOES NOT CURRENTLY WORK.
var tsProject = ts.createProject('./tsconfig.json');

// paths gotten from the config
var paths = {
	src: config.tasks.typescript.src,
	dest: config.tasks.typescript.dest
}

// meat of the task
var tsTask = function() {
	return browserify({ // we use browserify to bundle whatever so our application magically works. don't mess around with this
			basedir: "./",				// basically where this sees __dirname as 
			debug: true, 				// enable sourcemaps
			entries: [paths.src] 	// grabs all of our files from the root 
		})
		.plugin(tsify) // compile the typescript
		.bundle() // bundle up the result
		.pipe(source('app.js')) // the file output name
		.pipe(buffer()) // buffer so we can do the next step
		.pipe(sourcemaps.init({loadMaps: true})) // start up the sourcemap, load previously mapped maps
		// .pipe(uglify()) // minimize the code, does not work for whatever reason	
		.pipe(sourcemaps.write('.')) // write the sourcemap to the file system
		.pipe(gulp.dest(paths.dest));  // and output it to the destination in the config file
}
 
// the task defined with the variable above
gulp.task('ts', tsTask);

module.exports = tsTask