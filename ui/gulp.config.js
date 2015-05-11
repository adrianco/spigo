'use strict';

module.exports = {
	js: {
		src: ['./js/**/*.js', './js/**/*.json'],
		testsSrc: ['./tests/**/*.js'],
		entryFile: './js/app.js',
		outputFileName: 'bundle'
	},
	styles: {
		src: ['./styles/**/*.less'],
		entryFile: './styles/app.less',
		outputFileName: 'app'
	},
	dist: './dist',
	toCopy: {
		fontAwesomeFonts: './node_modules/font-awesome/fonts/*.*',
		googleFonts: './fonts/*.*',
		html: './index.html',
		json: '../json/*.json'
	}
};
