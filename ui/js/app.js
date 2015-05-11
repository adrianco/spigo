'use strict';

import Router from 'router';
import assign from 'lodash.assign';
import app from 'ampersand-app';

const router = new Router();

app.extend({
	init () {
		router.history.start();
	},

	navigate (path, opts = {}) {
		router.history.navigate(path, opts);
	}
});

app.init();

window.app = app;
