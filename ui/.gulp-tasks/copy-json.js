'use strict';

var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('copy-json', function() {
		return gulp
			.src(config.toCopy.json, { base: '../json' })
			.pipe(gulp.dest(config.dist + '/json'));
	});
};
