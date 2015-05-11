'use strict';

var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('copy-google-fonts', function() {
		return gulp
			.src(config.toCopy.googleFonts, { base: '.' })
			.pipe(gulp.dest(config.dist));
	});
};
