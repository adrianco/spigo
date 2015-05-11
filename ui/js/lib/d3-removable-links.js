'use strict';

import d3 from 'd3';

export default function() {
	return function(e) {
		d3.event.preventDefault();

		this.remove();
	};
};
