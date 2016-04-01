'use strict';

// config for Karma test runner for client tests
const gulpConfig = require('./gulp.config');

const files = [
	'./js/**/*.js',
	'./js/**/*.json',
	'./tests/**/*.js'
];

module.exports = function(config) {
	config.set({
		basePath: '',
		frameworks: ['browserify', 'jasmine'],
		files: files,
		exclude: [
			'./js/patches/**',
			'./js/app.js'
		],
		preprocessors: {
			'./js/**/*.js': 'browserify',
			'./js/**/*.json': 'browserify',
			'./tests/specs/**/*.js': 'browserify'
		},
		browserify: {
			debug: true,
			transform: [ 'babelify'],
			extensions: ['.js', '.json'],
			paths: ['./js']
		},
		browsers: ['PhantomJS'],
		reporters: ['nyan']
	});
};
