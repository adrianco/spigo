'use strict';

var uglify = require('gulp-uglify');
var plumber = require('gulp-plumber');
var rename = require('gulp-rename');
var notify = require('gulp-notify');
var connect = require('gulp-connect');
var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('minify-app', ['bundle-app'], function() {
		var src = config.dist + '/js/' + config.js.outputFileName + '.js';

		return gulp.src(src, { base: config.dist + '/js' })
			.pipe(plumber({
				errorHandler: notify.onError('Build Error: <%= error.message %>')
			}))
			.pipe(uglify())
			.pipe(rename(config.js.outputFileName + '.min.js'))
			.pipe(gulp.dest(config.dist + '/js'))
			.pipe(connect.reload())
			.pipe(notify('Client app minified...'));
	});
};
