var gulp   		= require('gulp');				// the task runner
var del    		= require('del');				// deletion of files/directories
var gutil		= require('gulp-util');				// logging

var config 		= require('../config.json');	// config file

var srcPaths = {
	// only need dest because that's what we are nuking
	dest: config.root.dest
}

var cleanTask = function (cb) {
  	return del(
        [srcPaths.dest],  // files to be deleted
        {
            dryRun: false,  // turn on if testing is needed
            force: true
        }
    ).then(paths => {
        gutil.log('Purged\n', paths.join('\n')); // log what was nuked
    });
}

gulp.task('clean', cleanTask)
module.exports = cleanTask