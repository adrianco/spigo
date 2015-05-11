'use strict';

import React from 'react';
import Header from 'header';

export default React.createClass({
	render () {
		return (
			<section id="app">
				<Header />
				{this.props.children}
			</section>
		);
	}
});
