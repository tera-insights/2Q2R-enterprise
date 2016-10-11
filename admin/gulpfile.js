var gulp = require('gulp');
var nodemon = require('gulp-nodemon');
var ts = require('gulp-typescript');
var concat = require('gulp-concat');
var sourcemaps = require('gulp-sourcemaps');

var print = require('gulp-print');

var destDir = '../server/assets/admin/';


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
    'libs/ui-router-extras/release/ct-ui-router-extras.min.js'
];

var libCssFiles = [
    'libs/angular-material/angular-material.min.css'
];


var tsProject = ts.createProject({
    module: "commonjs",
    target: "es5",
    preserveConstEnums: true,
    sourceMap: true
});

/**
 * We need to force a specifc order for typescript assets due to 
 * the fact that App.ts has to be last
 */

gulp.task('typescript', [], function () {
    var result = gulp.src([
            'src/controllers/**/*.ts',
            'src/interfaces/**/*.ts',
            'src/models/**/*.ts',
            'src/services/**/*.ts',
            'src/App.ts'
        ])
        .pipe(sourcemaps.init())
        .pipe(ts(tsProject));

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

gulp.task('build', ['typescript', 'copy assets', 'libs js', 'libs css']);
gulp.task('default', ['build', 'watch']);
