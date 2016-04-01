'use strict';

import React from 'react';
import $ from 'jquery';
import d3 from 'd3';
import reduce from 'lodash.reduce';
import each from 'lodash.foreach';
import bind from 'lodash.bind';
import tooltip from 'd3-tip';
import dispatcher from 'dispatcher';
import ChartStore from 'stores/chart';
import fisheye from 'lib/d3-fisheye';
import collideFactory from 'lib/d3-collision-detection';
import connectedNodesFactory from 'lib/d3-connected-nodes';
import removableNodesFactory from 'lib/d3-removable-nodes';
import removableLinksFactory from 'lib/d3-removable-links';
import linkExpanderFactory from 'lib/d3-link-expander';
import pinNodes from 'lib/d3-pin-nodes';

const fisheyeD3 = fisheye(d3);

const collide = collideFactory(fisheyeD3);

const HEADER_HEIGHT = 80;

const colorList = [
	'#1abc9c',
	'#2ecc71',
	'#3498db',
	'#9b59b6',
	'#34495e',
	'#f1c40f',
	'#e67e22',
	'#e74c3c',
	'#bdc3c7',
	'#7f8c8d',
	'#16a085'
];

const colors = (index) => {
	if (index < 0 || index > 10) index = Math.floor(Math.random() * 11);
	return colorList[index];
};

const pickColor = (d, i) => {
	var names = d[0].node.split('.');

	if (names.length < 4) return colors(0);
	else return colors(names[3].length);
};

export default React.createClass({
	getDefaultProps () {
		return {
			arch: 'migration',
			step: 0
		};
	},

	getInitialState () {
		return {
			width: window.innerWidth,
			height: window.innerHeight - HEADER_HEIGHT,
			charge: -1000
		};
	},

	updateSvgDims () {
		this.setState({
			width: window.innerWidth,
			height: window.innerHeight - HEADER_HEIGHT
		});
	},

	updateChart () {
		const dataset = ChartStore.getChartDataset();
		const {charge} = ChartStore.getStoreState();
		const {width, height} = this.state;

		if (!dataset.nodes.length) return;

		this.svg
			.selectAll('*')
			.remove();

		this.force
			.size([width, height])
			.nodes(dataset.nodes)
			.links(dataset.edges)
			.charge(charge)
			.linkDistance((d) => 10 + 7 * (d.source.size + d.target.size) / 2)
			.on('tick', bind(this._onTick, this))
			.start();

		this.links = this.svg
			.selectAll('.link')
			.data(dataset.edges)
			.enter()
			.append('line')
			.attr('class', 'link');

		this.nodes = this.svg
			.selectAll('.nodes')
			.data(dataset.nodes)
			.enter()
			.append('circle')
			.attr('class', 'node')
			.attr('r', (d) => Math.sqrt(d.size) * 2.6)
			.style('fill', pickColor)
			.call(pinNodes(fisheyeD3, this.force, bind(this._onTick, this)));

		const {mouseover, mouseout} = connectedNodesFactory(this.nodes, this.links);
		const removableNodes = removableNodesFactory(this.nodes, this.links);
		const removableLinks = removableLinksFactory();
		const {expand, shrink} = linkExpanderFactory();

		this.nodes
			.on('mouseover.connection', mouseover)
			.on('mouseout.connection', mouseout)
			.on('mouseover.tooltip', this.tip.show)
			.on('mouseout.tooltip', this.tip.hide)
			.on('dblclick', removableNodes);

		this.links
			.on('dblclick', removableLinks)
			.on('mouseover', expand)
			.on('mouseout', shrink);
	},

	componentWillMount () {
		this.boundUpdateSvgDims = bind(this.updateSvgDims, this);
	},

	componentDidMount () {
		const {arch, step} = this.props;

		this.svg = fisheyeD3.select(this.getDOMNode());
		this.force = fisheyeD3.layout.force();
		this.fisheye = fisheyeD3.fisheye
			.circular()
			.radius(230)
			.distortion(2);

		this.tip = tooltip(fisheyeD3)()
			.attr('class', 'd3-tip')
			.offset([-10, 0])
			.html((d) => d[0].node);

		this.svg.call(this.tip);

		setTimeout(() => {
			this.svg.on('mousemove', () => {
				this.force.stop();
				this.fisheye.focus(fisheyeD3.mouse(this.svg[0][0]));

				this.nodes
					.each(d => { d.fisheye = this.fisheye(d); })
					.attr('cx', d => d.fisheye.x)
					.attr('cy', d => d.fisheye.y)
					.attr('r', d => Math.sqrt(d.size) * 3);

				this.links
					.attr('x1', d => d.source.fisheye.x)
					.attr('y1', d => d.source.fisheye.y)
					.attr('x2', d => d.target.fisheye.x)
					.attr('y2', d => d.target.fisheye.y);
			});

			this.svg.on('mouseout', () => {
				this.force.resume();

				this.links
					.attr('x1', d => d.source.x)
					.attr('y1', d => d.source.y)
					.attr('x2', d => d.target.x)
					.attr('y2', d => d.target.y);

				this.nodes
					.attr('cx', (d) => d.x)
					.attr('cy', (d) => d.y)
					.attr('r', d => Math.sqrt(d.size) * 2.6);
			});
		}, 2000)

		ChartStore.addChangeListener(bind(this.updateChart), this);
		ChartStore.fetch(arch, step);
		window.addEventListener('resize', this.boundUpdateSvgDims);
	},

	componentDidUnmount () {
		ChartStore.removeChangeListener(bind(this.updateChart), this);
		window.removeEventListener('resize', this.boundUpdateSvgDims);
	},

	_onTick (d) {
		const {nodes} = ChartStore.getChartDataset();

		this.links
			.attr('x1', d => d.source.x)
			.attr('y1', d => d.source.y)
			.attr('x2', d => d.target.x)
			.attr('y2', d => d.target.y);

		this.nodes
			.attr('cx', (d) => d.x)
			.attr('cy', (d) => d.y)
			.each(collide(0.3, nodes));
	},

	render () {
		const {width, height} = this.state;

		return (<svg width={width} height={height}></svg>);
	}
});
