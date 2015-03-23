'use strict';

module.exports = function(gulp) {
	gulp.task('tdd', function() {
		gulp.watch([
			'./js/**/*.js',
			'./js/**/*.json',
			'./tests/**/*.js'
		], ['lint', 'test']);
	});
};
