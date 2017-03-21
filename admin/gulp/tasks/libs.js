/*
    Get all of our libs and concat them into their own things
*/
var gulp = require('gulp'); // the gulp runner
var concat = require('gulp-concat'); // concatenate assets into one file for less http requests
var gutil = require('gulp-util'); // logging of lib files

var config = require('../config.json'); // the config file

// javascript lib files
var libJsFiles = [
    // path from the gulpfile
    'node_modules/jquery/dist/jquery.min.js',
    'node_modules/random-js/lib/random.min.js',
    'node_modules/prob.js/dist/prob-min.js',
    'node_modules/angular/angular.min.js',
    'node_modules/angular-aria/angular-aria.min.js',
    'node_modules/angular-material/angular-material.min.js',
    'node_modules/angular-resource/angular-resource.min.js',
    'node_modules/angular-messages/angular-messages.min.js',
    'node_modules/angular-secure-password/dist/angular-secure-password.js',
    'node_modules/angular-ui-router/release/angular-ui-router.min.js',
    'node_modules/angular-animate/angular-animate.min.js',
    'node_modules/ui-router-extras/release/ct-ui-router-extras.min.js',
    'node_modules/angular-material-data-table/dist/md-data-table.min.js',
    'node_modules/ng-file-upload/dist/ng-file-upload.min.js',
    'node_modules/leaflet/dist/leaflet.js',
    'node_modules/chart.js/dist/Chart.bundle.min.js'
];

// css lib files
var libCssFiles = [
    'node_modules/material-design-icons/iconfont/material-icons.css', // old iconset
    'node_modules/angular-material/angular-material.min.css', // angular material
    'node_modules/angular-material-data-table/dist/md-data-table.min.css', // the table we use
    'node_modules/leaflet/dist/leaflet.css', // maps
    'node_modules/angular-secure-password/dist/angular-secure-password.css', // our library, for secure password input
    'node_modules/mdi/css/materialdesignicons.min.css' // new iconset
];

// fonts, mostly iconsets right now
var libFontFiles = [
    // the iconset webfonts
    'node_modules/mdi/fonts/materialdesignicons-webfont.woff',
    'node_modules/mdi/fonts/materialdesignicons-webfont.ttf',
    'node_modules/mdi/fonts/materialdesignicons-webfont.eot'
];

// javascripts for installed packages
gulp.task('libs js', function() {
    return gulp.src(libJsFiles)
        .pipe(concat('libs.min.js'))
        .pipe(gulp.dest(config.tasks.typescript.dest));
});

// for testing, trying to see if something is going wrong
gulp.task('log js', function() {
    // dump all of the lib files to be loaded in the console
    return libJsFiles.forEach(function(libJsFiles) {
        gutil.log(libJsFiles);
    })
});


// styles from installed packages
gulp.task('libs css', function() {
    return gulp.src(libCssFiles)
        .pipe(concat('libs.min.css'))
        .pipe(gulp.dest(config.tasks.less.dest));
});

// fonts from installed packages
gulp.task('libs font', function() {
    return gulp.src(libFontFiles)
        // make sure to pipe into a fonts directory
        .pipe(gulp.dest(config.tasks.fonts.dest));
});
