'use strict';

import React from 'react';
import Router from 'ampersand-router';
import SimianViz from 'simianviz';
import FourOhFour from 'four-oh-four';
import indexOf from 'lodash.indexof';

const archs = [
	'fsm',
	'migration',
	'netflixoss',
	'lamp',
	'container',
	'aws_ac_ra_web',
	'netflix',
	'cassandra',
	'yogi',
	'riak',
];

export default Router.extend({
	routes: {
		'': 'default',
		':arch(/:step)': 'deepLink',
		'*404': 'fourOhFour'
	},

	default () {
		React.render(<SimianViz />, document.body);
	},

	deepLink (arch, step) {
		if (!~indexOf(archs, arch)) return this.fourOhFour();
		React.render(<SimianViz arch={arch} step={step} />, document.body);
	},

	fourOhFour () {
		React.render(<FourOhFour />, document.body);
	}
});
