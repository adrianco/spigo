'use strict';

var clean = require('gulp-clean');
var notify = require('gulp-notify');
var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('clean-dist', function () {
		return gulp.src(config.dist, { read: false })
			.pipe(clean({ force: true }))
			.pipe(notify('Client dist folder cleaned...'));
	});
};
