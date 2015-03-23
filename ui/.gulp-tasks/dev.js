'use strict';

var connect = require('gulp-connect');
var runSequence = require('run-sequence');
var config = require('../gulp.config.js');

module.exports = function(gulp) {
	gulp.task('dev', function(done) {
		gulp.watch(config.styles.src, ['styles'], function() {
			connect.reload();
		});

		gulp.watch(config.js.src, ['bundle-app'], function() {
			connect.reload();
		});

		runSequence(['build-dev'], 'serve', done);
	});
};
