'use strict';

var runSequence = require('run-sequence');

module.exports = function(gulp) {
	gulp.task('build-prod', function(done) {
		runSequence('clean-dist', [
			'lint',
			'minify-styles',
			'copy-font-awesome-fonts',
			'copy-google-fonts',
			'copy-html',
			'copy-json',
			'minify-app'
		], done);
	});
};
