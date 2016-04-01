'use strict';

var plumber = require('gulp-plumber');
var rename = require('gulp-rename');
var connect = require('gulp-connect');
var notify = require('gulp-notify');
var minifyCss = require('gulp-clean-css');
var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('minify-styles', ['styles'], function() {
		var src = config.dist + '/css/' + config.styles.outputFileName + '.css';

		return gulp.src(src)
			.pipe(plumber({
				errorHandler: notify.onError('Build Error: <%= error.message %>')
			}))
			.pipe(minifyCss())
			.pipe(rename(config.styles.outputFileName + '.min.css'))
			.pipe(gulp.dest(config.dist + '/css/'))
			.pipe(connect.reload())
			.pipe(notify('Styles created...'));
	});
};
