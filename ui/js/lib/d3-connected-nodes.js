'use strict';

import each from 'lodash.foreach';
import d3 from 'd3';

export default function(nodes, links) {

	//Create an array logging what is connected to what
	let linkedByIndex = {};

	each(nodes.data(), (n, i) => {
		linkedByIndex[i + ',' + i] = 1;
	});

	each(links.data(), (l) => {
		linkedByIndex[l.source.index + ',' + l.target.index] = 1;
	});

	//This function looks up whether a pair are neighbours
	function neighboring(a, b) {
		return linkedByIndex[a.index + ',' + b.index];
	}

	return {
		mouseover () {
			const d = d3.select(this).node().__data__;

			nodes
				.transition()
				.duration(100)
				.style('opacity', (o) => {
					return neighboring(d, o) | neighboring(o, d) ? 1 : 0.1;
				});

			links
				.transition()
				.duration(100)
				.style('opacity', function (o) {
					return d.index === o.source.index | d.index === o.target.index ? 1 : 0.1;
				});
		},

		mouseout () {
			nodes
				.transition()
				.duration(300)
				.style('opacity', 1);
			links
				.transition()
				.duration(300)
				.style('opacity', 1);
		}
	};
};
