'use strict';

var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('copy-font-awesome-fonts', function() {
		return gulp
			.src(config.toCopy.fontAwesomeFonts, { base: './node_modules/font-awesome' })
			.pipe(gulp.dest(config.dist));
	});
};
