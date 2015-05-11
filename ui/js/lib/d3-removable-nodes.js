'use strict';

import d3 from 'd3';
import filter from 'lodash.filter';
import each from 'lodash.foreach';

export default function(nodes, links) {

	function removeLinks(nodeName) {
		let sources = filter(links[0], (l) => {
			return l.__data__.source[0].node === nodeName;
		});
		let targets = filter(links[0], (l) => {
			return l.__data__.target[0].node === nodeName;
		});

		each([].concat(sources, targets), (l) => l.remove());
	}

	return function(e) {
		d3.event.preventDefault();

		const d = d3.select(this).node().__data__;

		removeLinks(d[0].node);
		this.remove();
	};
};
