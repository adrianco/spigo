'use strict';

const PADDING = 1;

export default (d3) => {
	return (alpha, nodes) => {
		var quadtree = d3.geom.quadtree(nodes);

		return function(d) {
			var rb = 2 * 10 + PADDING,
				nx1 = d.x - rb,
				nx2 = d.x + rb,
				ny1 = d.y - rb,
				ny2 = d.y + rb;

			quadtree.visit(function(quad, x1, y1, x2, y2) {
				if (quad.point && (quad.point !== d)) {
					var x = d.x - quad.point.x,
						y = d.y - quad.point.y,
						l = Math.sqrt(x * x + y * y);

					if (l < rb) {
						l = (l - rb) / l * alpha;
						d.x -= x *= l;
						d.y -= y *= l;
						quad.point.x += x;
						quad.point.y += y;
					}
				}

				return x1 > nx2 ||
						x2 < nx1 ||
						y1 > ny2 ||
						y2 < ny1;
			});
		};
	};
};
