'use strict';

var plumber = require('gulp-plumber');
var less = require('gulp-less');
var rename = require('gulp-rename');
var connect = require('gulp-connect');
var notify = require('gulp-notify');
var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('styles', function() {
		return gulp.src(config.styles.entryFile )
			.pipe(plumber({
				errorHandler: notify.onError("Build Error: <%= error.message %>")
			}))
			.pipe(less())
			.pipe(rename(config.styles.outputFileName + '.css'))
			.pipe(gulp.dest(config.dist + '/css/'))
			.pipe(connect.reload())
			.pipe(notify('Styles created...'));
	});
};
