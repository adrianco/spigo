'use strict';

var runSequence = require('run-sequence');

module.exports = function(gulp) {
	gulp.task('build-dev', function(done) {
		runSequence('clean-dist', [
			'lint',
			'styles',
			'copy-font-awesome-fonts',
			'copy-google-fonts',
			'copy-html',
			'copy-json',
			'bundle-app'
		], done);
	});
};
