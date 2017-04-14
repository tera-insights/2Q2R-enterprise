var gulp            = require('gulp')
var gulpSequence    = require('gulp-sequence')

var defaultTask = function(cb) {
  	// smashes all the tasks together, cb has to be at the end
  	gulpSequence('clean', 'less', 'html', 'images', 'libs css', 'log js', 'libs js', 'libs font', 'tslint', 'ts', 'watch', cb);
}

// sets the default task
gulp.task('default', defaultTask);
module.exports = defaultTask