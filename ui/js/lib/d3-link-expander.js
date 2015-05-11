'use strict';

import d3 from 'd3';

export default function() {
	return {

		expand () {
			const l = d3.select(this);

			l.style('stroke-width', '4px')
				.style('stroke', '#c0392b');
		},

		shrink () {
			const l = d3.select(this);

			l.style('stroke-width', '1.5px')
				.style('stroke', "#999");
		}
	};
};
