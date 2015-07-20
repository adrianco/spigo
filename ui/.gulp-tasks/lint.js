'use strict';

var eslint = require('gulp-eslint');
var plumber = require('gulp-plumber');
var notify = require('gulp-notify');
var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('lint', function() {
		var src = [].concat(
			config.js.src,
			[
				'./gulpfile.js',
				'./.gulp-tasks/*.js',
				'!./**/*.json'
			]
		);

		return gulp
			.src(src)
			.pipe(plumber({
				errorHandler: notify.onError('Build Error: <%= error.message %>')
			}))
			.pipe(eslint())
			.pipe(eslint.format())
			.pipe(eslint.failOnError());
	});
};
