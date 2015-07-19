'use strict';

import eventMixin from 'lib/store-event-mixin';
import actions from 'actions';
import assign from 'lodash.assign';
import each from 'lodash.foreach';
import pluck from 'lodash.pluck';
import filter from 'lodash.filter';
import bind from 'lodash.bind';
import sortBy from 'lodash.sortby';
import apiClient from 'api-client';
import Promise from 'bluebird';
import dispatcher from 'dispatcher';
import app from 'ampersand-app';

const ChartStore = assign({}, eventMixin, {

	CHANGE_EVENT: actions.CHART_DATA_CHANGED,

	cache: {},

	charge: -400,

	updateCharge (newValue) {
		this.charge = newValue;
		this.emitChange();
	},

	fetchJSON () {
		let {architecture, step} = this;

		if (!architecture) console.error('Architecture needed to fetch JSON!');
		if (this.cache[`${architecture}${step}`]) return Promise.resolve();

		step = (step === 0) ? '' : step;

		return new Promise((res, rej) => {
			apiClient
				.get(`/json/${architecture}${step}.json`)
				.then((data) => {
					this.processJSON(data);
					res();
				})
				.catch((err) => {
					console.error(err);
					rej();
				});
		});
	},

	processJSON (data, step) {
		if (step === undefined) step = this.step;

		let nodes = [], edges = [], nodeCount = 0;

		const addNode = function (name) {
			nodes.push([name]);
			nodes[nodeCount].size = 1;
			nodeCount++;
		};

		const findNode = function (name) {
			var found = false,
					index = 0;

			while (!found)
				if (nodes[index] &&
					nodes[index][0].node === name) found = true;
				else index++;

			nodes[index].size++;
			return nodes[index];
		};

		const addEdge = function (source, target) {
			var sn = findNode(source);
			var tn = findNode(target);
			edges.push({ source: sn, target: tn });
		};

		const unprocessedNodes = filter(data.body.graph, (e) => !!e.node);
		const unprocessedEdges = filter(data.body.graph, (e) => !!e.edge);

		each(unprocessedNodes, (n) => addNode(n));
		each(unprocessedEdges, (e) => addEdge(e.source, e.target));

		nodes = sortBy(nodes, 'timestamp');
		edges = sortBy(edges, 'timestamp');

		this.cache[this.architecture + step] = {nodes, edges};
	},

	fetch (arch, step) {
		if (arch === undefined) return console.error('Architecture needed to fetch JSON!');

		this.architecture = arch;
		this.step = (!step) ? 0 : parseInt(step, 10);

		this.fetchJSON()
			.then(bind(this.prefetchNextStep, this))
			.then(() => {
				const step = (this.step) ? ('/' + this.step) : '';
				const path = `/${this.architecture}${step}`;

				this.emitChange();
				app.navigate(path, { trigger: false });
			})
			.error(console.error);
	},

	prev () {
		this.fetch(this.architecture, this.step - 1);
	},

	next () {
		this.fetch(this.architecture, this.step + 1);
	},

	prefetchNextStep () {
		let {architecture, step} = this;
		step = step + 1;

		if (this.cache[architecture + step]) return Promise.resolve();

		return new Promise((res, rej) => {
			apiClient
				.get(`/json/${architecture}${step}.json`)
				.then((data) => {
					if (data.body) this.processJSON(data, step);
					res();
				})
				.catch((err) => {
					console.error(err);
					rej();
				});
		});
	},

	getStoreState () {
		const isCached = (this.cache[this.architecture + (this.step + 1)]) ?
				true : false;

		return {
			hasNextStep: isCached,
			hasPreviousStep: this.step > 0,
			architecture: this.architecture,
			charge: this.charge
		};
	},

	getChartDataset () {
		if (!this.cache[this.architecture + this.step])
			return { nodes: [], edges: [] };

		const {nodes, edges} = this.cache[this.architecture + this.step];

		return {
			nodes: nodes,
			edges: edges
		};
	}
});

ChartStore.dispatchToken = dispatcher.register((event) => {
	switch (event.type) {
		case actions.FETCH_PREVIOUS_CHART:
			ChartStore.prev();
			break;
		case actions.FETCH_NEXT_CHART:
			ChartStore.next();
			break;
		case actions.ARCHITECTURE_UPDATED:
			ChartStore.fetch(event.data);
			break;
		case actions.CHARGE_UPDATED:
			ChartStore.updateCharge(event.data);
			break;
	}
});

export default ChartStore;
