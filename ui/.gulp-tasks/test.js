'use strict';

var karma = require('karma').server;
var path = require('path');

module.exports = function(gulp) {
	gulp.task('test', ['lint'], function (done) {

		karma.start({
			configFile: path.resolve(__dirname, '../karma.config.js'),
			singleRun: true
		}, done);
	});
};
