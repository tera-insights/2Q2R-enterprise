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
//var destDir = '../server/assets/admin/';
var destDir = path.join(__dirname, "..", "server", "assets", "admin");

/**
 * Add all the bower assets to these lists so they can be concatenated 
 * into the assets directory. When this is changed, restart gulp
 */
var libJsFiles = [
    'libs/jquery/dist/jquery.min.js',
    'libs/angular/angular.min.js',
    'libs/angular-animate/angular-animate.min.js',
    'libs/angular-material/angular-material.min.js',
    'libs/angular-aria/angular-aria.min.js',
    'libs/angular-resource/angular-resource.min.js',
    'libs/angular-messages/angular-messages.min.js',
    'libs/angular-ui-router/release/angular-ui-router.min.js',
    'libs/ui-router-extras/release/ct-ui-router-extras.min.js',
    'libs/lodash/dist/lodash.min.js',
    'libs/random/lib/random.min.js',
    'libs/prob.js/dist/prob-min.js',
    'libs/angular-sanitize/angular-sanitize.min.js',
    'libs/angular-material-icons/angular-material-icons.min.js',
    'libs/md-data-table/dist/md-data-table-templates.js',
    'libs/md-data-table/dist/md-data-table.js'
];

var libCssFiles = [
    'libs/material-design-icons/iconfont/material-icons.css',
    'libs/angular-material/angular-material.min.css',
    'libs/angular-material-icons/angular-material-icons.css',
    'libs/md-data-table/dist/md-data-table-style.css'
];

var libFontFiles = [
    'libs/material-design-icons/iconfont/MaterialIcons-Regular.woff2',
    'libs/material-design-icons/iconfont/MaterialIcons-Regular.woff',
    'libs/material-design-icons/iconfont/MaterialIcons-Regular.ttf',
];


var tsProject = ts.createProject({
    module: "commonjs",
    target: "es5",
    removeComments: true,
    preserveConstEnums: true,
    sourceMap: true
});


// Creates one large javascript file 
gulp.task("typescript", function () {
    var bundler = browserify({ basedir: "./" })
        .add(path.join("./src", "App.ts"))
        .plugin(tsify, tsProject);
    return bundler.bundle()
        .on("error", function (error) {
            console.log(error.toString());
        })
        .pipe(source('app.min.js'))
        .pipe(gulp.dest(path.join(destDir,"js")));
});

/**
 * We need to force a specifc order for typescript assets due to 
 * the fact that App.ts has to be last
 */

gulp.task('ttypescript', [], function () {
    var result = gulp.src([
        'src/controllers/**/*.ts',
        'src/interfaces/**/*.ts',
        'src/models/**/*.ts',
        'src/services/**/*.ts',
        'src/App.ts'
    ])
        .pipe(sourcemaps.init())
        .pipe(tsProject());

    return result.js
        .pipe(sourcemaps.write())
        .pipe(print())
        .pipe(concat('app.min.js'))
        .pipe(gulp.dest(destDir + '/js'));
});

gulp.task('libs js', function () {
    return gulp.src(libJsFiles)
        .pipe(concat('libs.min.js'))
        .pipe(gulp.dest(destDir + '/js'));
});

gulp.task('libs css', function () {
    return gulp.src(libCssFiles)
        .pipe(concat('libs.min.css'))
        .pipe(gulp.dest(destDir + '/css'));
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
