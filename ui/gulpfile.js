'use strict';

var gulp		= require('gulp');
var shell		= require('shelljs');

shell.ls('.gulp-tasks')
	.filter(function(file) { return file.match(/\.js$/); })
	.map(function(file) { return file.replace('.js', ''); })
	.forEach(function(fileName) { require('./.gulp-tasks/' + fileName)(gulp); });
