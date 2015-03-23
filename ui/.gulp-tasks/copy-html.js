'use strict';

var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('copy-html', function() {
		return gulp
			.src(config.toCopy.html, { base: '.' })
			.pipe(gulp.dest(config.dist));
	});
};
