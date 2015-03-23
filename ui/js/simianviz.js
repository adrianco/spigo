'use strict';

import React from 'react';
import Layout from 'layout';
import Toolbar from 'toolbar';
import Chart from 'chart';

export default React.createClass({
	render () {
		return (
			<Layout>
				<Toolbar {...this.props} />
				<Chart {...this.props} />
			</Layout>
		);
	}
});
