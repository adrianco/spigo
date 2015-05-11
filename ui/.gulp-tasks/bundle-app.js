'use strict';

var plumber = require('gulp-plumber');
var notify = require('gulp-notify');
var browserify = require('browserify');
var source = require('vinyl-source-stream');
var babelify = require('babelify');
var connect = require('gulp-connect');
var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('bundle-app', function() {
		return browserify({
				entries: config.js.entryFile,
				extensions: ['.js', '.json'],
				paths: ['./js'],
				debug: true
			})
			.transform(babelify)
			.bundle()
			.on('error', function(err){
				console.log(err.message);
			})
			.pipe(plumber({
				errorHandler: notify.onError("Build Error: <%= error.message %>")
			}))
			.pipe(source(config.js.outputFileName + '.js'))
			.pipe(gulp.dest(config.dist + '/js'))
			.pipe(connect.reload())
			.pipe(notify('Client app bundled...'));
	});
};
