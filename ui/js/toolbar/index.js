'use strict';

import React from 'react';
import ChartStore from 'stores/chart';
import bind from 'lodash.bind';
import dispatcher from 'dispatcher';
import actions from 'actions';

export default React.createClass({
	getInitialState () {
		return ChartStore.getStoreState();
	},

	componentWillMount () {
		this.boundOnChange = bind(this._onChange, this);
	},

	componentDidMount () {
		ChartStore.addChangeListener(this.boundOnChange);
	},

	componentWillUnmount () {
		ChartStore.removeChangeListener(this.boundOnChange);
	},

	_onChange () {
		this.setState(ChartStore.getStoreState());
	},

	prev (e) {
		dispatcher.dispatch({ type: actions.FETCH_PREVIOUS_CHART });
	},

	next (e) {
		dispatcher.dispatch({ type: actions.FETCH_NEXT_CHART });
	},

	arch (e) {
		this.setState({ architecture: e.target.value });

		dispatcher.dispatch({
			type: actions.ARCHITECTURE_UPDATED,
			data: e.target.value
		});
	},

	charge (e) {
		const charge = parseInt(e.target.value, 10);
		this.setState({ charge: charge });

		dispatcher.dispatch({
			type: actions.CHARGE_UPDATED,
			data: charge
		});
	},

	render () {
		const {hasNextStep, hasPreviousStep, architecture, charge} = this.state;
		let prev, next;

		if (hasPreviousStep)
			prev = <span
				id="prev"
				className="action"
				title="Previous Step"
				onClick={this.prev}>

				<i className="fa fa-angle-left"></i>
			</span>;
		else
			prev = null;

		if (hasNextStep)
			next = <span
				id="next"
				className="action"
				title="Next Step"
				onClick={this.next}>

				<i className="fa fa-angle-right"></i>
			</span>;
		else
			next = null;

		return (
			<section id="toolbar">
				{prev}{next}

				<span id="charge-button" className="action" title="Add Charge">
					Charge:
					<select
						name="charge"
						id="charge"
						value={charge}
						onChange={this.charge}>

						<option value="-200">200</option>
						<option value="-400">400</option>
						<option value="-600">600</option>
						<option value="-800">800</option>
						<option value="-1000">1000</option>
					</select>
				</span>

				<span className="action" title="Architecture">
					Architecture:
					<select
						name="architecture"
						id="architecture"
						value={architecture}
						onChange={this.arch}>

						<option value="lamp">Lamp</option>
						<option value="netflixoss">Netflix OSS</option>
						<option value="fsm">FSM</option>
						<option value="migration">Migration</option>
						<option value="container">Container</option>
						<option value="aws_ac_ra_web">AWS Reference Arch</option>
						<option value="netflix">More complex Netflix</option>
						<option value="yogi">Contributed data science</option>
					</select>
				</span>
			</section>
		);
	}
});
