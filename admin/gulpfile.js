var gulp = require('gulp');
var nodemon = require('gulp-nodemon');
var ts = require('gulp-typescript');
var concat = require('gulp-concat');
var sourcemaps = require('gulp-sourcemaps');
var browserify = require("browserify");
var tsify = require("tsify");
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
    'node_modules/angular-secure-password/dist/angular-secure-password.js'
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


// Creates one large javascript file 
gulp.task("typescript", function () {
    var bundler = browserify({ basedir: "./" })
        .add(path.join("./src", "App.ts"))
        .plugin(tsify, tsProject);
    return bundler.bundle()
        .on("error", function (error) {
            console.log(error.message, error.fileName, error.lineNumber);
        })
        .pipe(source('app.min.js'))
        .pipe(gulp.dest(path.join(destDir,"js")));
});

gulp.task('libs js', function () {
    return gulp.src(libJsFiles)
        .pipe(concat('libs.min.js'))
        .pipe(gulp.dest(path.join(destDir, 'js')));
});

gulp.task('libs css', function () {
    return gulp.src(libCssFiles)
        .pipe(concat('libs.min.css'))
        .pipe(gulp.dest(path.join(destDir, 'css')));
});

gulp.task('libs font', function () {
    return gulp.src(libFontFiles)
        .pipe(gulp.dest(destDir + '/css'));
});

gulp.task('copy assets', function () {
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

gulp.task('watch', function () {
    gulp.watch('src/**/*.ts', ['typescript']);
    gulp.watch(['src/**/*.*', '!src/**/*.ts'], ['copy assets']);
});

gulp.task('build', ['typescript', 'copy assets', 'libs js', 'libs css', 'libs font']);
gulp.task('default', ['build', 'watch']);