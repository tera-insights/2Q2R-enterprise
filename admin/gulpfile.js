var gulp = require('gulp'); // the gulp runner
var nodemon = require('gulp-nodemon'); // gulp port of nodemon, monitors changes
var ts = require('gulp-typescript'); // typescript compiler
var concat = require('gulp-concat'); // concatenate assets into one file for less http requests
var sourcemaps = require('gulp-sourcemaps'); // sourcemaps to debug typescript with the dev console
var browserify = require("browserify");
var tsify = require("tsify");
var minifyCSS = require('gulp-minify-css'); // minify css before concat
var less = require('gulp-less'); // less compiler
var path = require("path");
var print = require('gulp-print');
var source = require("vinyl-source-stream");
var destDir = path.join(__dirname, "..", "server", "assets", "admin");

/**
 * Add all the bower assets to these lists so they can be concatenated 
 * into the assets directory. When this is changed, restart gulp
 */
var libJsFiles = [
    'node_modules/jquery/dist/jquery.min.js',
    'node_modules/random-js/lib/random.min.js',
    'node_modules/prob.js/dist/prob-min.js',
    'node_modules/angular/angular.min.js',
    'node_modules/angular-animate/angular-animate.min.js',
    'node_modules/angular-material/angular-material.min.js',
    'node_modules/angular-aria/angular-aria.min.js',
    'node_modules/angular-resource/angular-resource.min.js',
    'node_modules/angular-messages/angular-messages.min.js',
    'node_modules/angular-ui-router/release/angular-ui-router.min.js',
    'node_modules/ui-router-extras/release/ct-ui-router-extras.min.js',
    'node_modules/angular-material-data-table/dist/md-data-table.min.js',
    'node_modules/leaflet/dist/leaflet.js',
    'node_modules/chart.js/dist/Chart.bundle.min.js',
    'node_modules/angular-secure-password/dist/angular-secure-password.js',
    'node_modules/ng-file-upload/dist/ng-file-upload.min.js'
];

var libCssFiles = [
    'node_modules/material-design-icons/iconfont/material-icons.css',
    'node_modules/angular-material/angular-material.min.css',
    'node_modules/angular-material-data-table/dist/md-data-table.min.css',
    'node_modules/leaflet/dist/leaflet.css',
    'node_modules/angular-secure-password/dist/angular-secure-password.css'
];

var libFontFiles = [
    'node_modules/material-design-icons/iconfont/MaterialIcons-Regular.woff2',
    'node_modules/material-design-icons/iconfont/MaterialIcons-Regular.woff',
    'node_modules/material-design-icons/iconfont/MaterialIcons-Regular.ttf'
];


var tsProject = ts.createProject({
    module: "commonjs",
    target: "es5",
    removeComments: true,
    preserveConstEnums: true,
    sourceMap: true,
    lib: [
        "ES2015.Promise"
    ]
});

// less files in the application, concat them


// find and concatenate less files, usually within the 'css' folder
gulp.task("less", function() {
    gulp.src(['src/**/*.less'])
        .pipe(less()) // less compiler
        .pipe(minifyCSS()) // minify css files returned
        .pipe(concat('styles.min.css')) // concat them
        .pipe(gulp.dest('src/css/')) // and send back to /src/css/
});

// Creates one large javascript file 
gulp.task("typescript", function() {
    var bundler = browserify({ basedir: "./" })
        .add(path.join("./src", "App.ts"))
        .plugin(tsify, tsProject);
    return bundler.bundle()
        .on("error", function(error) {
            console.log(error.message, error.fileName, error.lineNumber);
        })
        .pipe(source('app.min.js'))
        .pipe(gulp.dest(path.join(destDir, "js")));
});

// javascripts for installed packages
gulp.task('libs js', function() {
    return gulp.src(libJsFiles)
        .pipe(concat('libs.min.js'))
        .pipe(gulp.dest(path.join(destDir, 'js')));
});

// styles from installed packages
gulp.task('libs css', function() {
    return gulp.src(libCssFiles)
        .pipe(concat('libs.min.css'))
        .pipe(gulp.dest(path.join(destDir, 'css')));
});

// fonts from installed packages
gulp.task('libs font', function() {
    return gulp.src(libFontFiles)
        .pipe(gulp.dest(destDir + '/css'));
});

gulp.task('copy assets', function() {
    // Also, copy over other assets
    var result = gulp.src([
        'src/**/*.css',
        'src/**/*.js',
        'src/**/*.html',
        'src/**/*.png',
        'src/**/*.jpg',
        'src/**/*.svg',
        'src/**/*.ico'
    ]);

    return result.pipe(gulp.dest(destDir));
});

gulp.task('watch', function() {
    gulp.watch('src/**/*.ts', ['typescript']); // watch for changes in typescript files
    gulp.watch('src/**/*.less', ['less']); // in less files
    gulp.watch(['src/**/*.*', '!src/**/*.ts'], ['copy assets']); // in every other file
});

gulp.task('build', ['typescript', 'copy assets', 'libs js', 'libs css', 'less', 'libs font']);
gulp.task('default', ['build', 'watch']);
