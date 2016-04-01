'use strict';

var rename = require('gulp-rename');

var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('copy-html', function() {
		return gulp
			.src(config.toCopy.html, { base: '.' })
			.pipe(gulp.dest(config.dist))
			.pipe(rename('200.html'))
			.pipe(gulp.dest(config.dist));
	});
};
