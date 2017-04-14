/* 
	How this works
	I used gulp-starter as a base to make things easier;
		- NPM is the package manager
		- NPM is also the typings manager
		- Typescript is a thing
		- And a bunch of other stuff listed in the git repo for angular2-starter, might write down (???)

	This gulpfile is just a base for the more organized file-based system in the gulp directory; it just pulls the requesite files
	Adding a new task means adding a new js file to the gulp/tasks directory	
*/

var gulp 			= require('gulp'); 						// the task runner, gulp; always needs to be first
var	requireDir 		= require('require-dir'); 				// require specific directories within our file system
var config 			= require('./gulp/config.json');		// use a config file for everything, call with config.whatever

// recursively get all the files inside of gulp/tasks
requireDir('./gulp/tasks/');

