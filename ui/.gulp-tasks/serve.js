'use strict';

var connect = require('gulp-connect');

module.exports = function(gulp) {
	gulp.task('serve', function() {

		gulp.task('serve', function () {
			connect.server({
				root: ['dist'],
				port: 8000,
				livereload: true,
				fallback: 'index.html'
			});
		});
	});
};
