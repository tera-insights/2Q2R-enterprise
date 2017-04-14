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
    'node_modules/jquery/dist/jquery.min.js',                                   // used for miscellaneous purposes, is a dependency of basically everything, etc etc.
    'node_modules/random-js/lib/random.min.js',                                 // for random random numbers, used in some encrytion things
    'node_modules/prob.js/dist/prob-min.js',                                    // also a random number generator (???)   
    'node_modules/angular/angular.min.js',                                      // main framework we use
    'node_modules/angular-aria/angular-aria.min.js',                            // for ARIA attributes (screenreaders) and a dependency of angular material
    'node_modules/angular-material/angular-material.min.js',                    // our ui framework, the majority of ui elements on the screen are from here
    'node_modules/angular-resource/angular-resource.min.js',                    // interacts with REST server data sources
    'node_modules/angular-messages/angular-messages.min.js',                    // used for form validation, also a dependency of angular material
    'node_modules/angular-secure-password/dist/angular-secure-password.js',     // our library that we use for custom password inputs, fix up someday :(
    'node_modules/angular-ui-router/release/angular-ui-router.min.js',          // manages states in our application
    'node_modules/angular-animate/angular-animate.min.js',                      // animations via angular, dependency of angular material
    'node_modules/ui-router-extras/release/ct-ui-router-extras.min.js',         // adds some nifty things to ui router
    'node_modules/angular-material-data-table/dist/md-data-table.min.js',       // our table library, follows material design spec
    'node_modules/ng-file-upload/dist/ng-file-upload.min.js',                   // used for file uploading and drag and drop management, probably other things
    'node_modules/leaflet/dist/leaflet.js',                                     // used for maps
    'node_modules/chart.js/dist/Chart.bundle.min.js',                           // current charting library, will probably be replaced with ours in the future
    'node_modules/tinycolor2/dist/tinycolor-min.js',                            // dependency of color picker below     
    'node_modules/angularjs-color-picker/dist/angularjs-color-picker.min.js'    // color picker     
];

// css lib files
var libCssFiles = [
    'node_modules/angular-material/angular-material.min.css',                   // angular material
    'node_modules/angular-material-data-table/dist/md-data-table.min.css',      // the table we use
    'node_modules/leaflet/dist/leaflet.css',                                    // maps
    'node_modules/angular-secure-password/dist/angular-secure-password.css',    // our library for secure password input
    'node_modules/mdi/css/materialdesignicons.min.css',                         // material design iconset
    'node_modules/angularjs-color-picker/dist/angularjs-color-picker.min.css'   // styles for color picker
];  

// fonts, mostly iconsets right now
var libFontFiles = [
    // the iconset webfonts
    'node_modules/mdi/fonts/materialdesignicons-webfont.woff',
    'node_modules/mdi/fonts/materialdesignicons-webfont.ttf',
    'node_modules/mdi/fonts/materialdesignicons-webfont.eot',
    'node_modules/roboto-fontface/fonts/Roboto/Roboto-Regular.ttf',             // The normal roboto font weight
    'node_modules/roboto-fontface/fonts/Roboto/Roboto-Regular.woff',            // ^^
    'node_modules/roboto-fontface/fonts/Roboto/Roboto-Regular.woff2',           // ^^
    'node_modules/roboto-fontface/fonts/Roboto/Roboto-Medium.ttf',              // Medium weight
    'node_modules/roboto-fontface/fonts/Roboto/Roboto-Medium.woff',             // ^^
    'node_modules/roboto-fontface/fonts/Roboto/Roboto-Medium.woff2',            // ^^
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
