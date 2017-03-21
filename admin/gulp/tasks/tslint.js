var gulp 			= require('gulp');					// the task runner
var tslint			= require('gulp-tslint');			// linter for typescript

var config 			= require('../config.json');		// config file that controls src/dest (mostly)

// paths gotten from the config
var paths = {
	src: config.tasks.typescript.src,
	dest: config.tasks.typescript.dest
}


// linter, check ts files before smashing them together 
var lintTask = function(){
      gulp.src(paths.src)
        .pipe(tslint({
        	configuration: "./tslint.json", // look for the tslint file
        	formatter: "prose"
        }))
        .pipe(tslint.report());
}

// lint task
gulp.task('tslint', lintTask);

module.exports = lintTask;