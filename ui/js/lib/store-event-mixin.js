'use strict';

import {EventEmitter} from 'events';
import assign from 'lodash.assign';

export default assign({}, EventEmitter.prototype, {
	emitChange () {
		this.emit(this.CHANGE_EVENT);
	},

	addChangeListener (fn) {
		this.on(this.CHANGE_EVENT, fn);
	},

	removeChangeListener (fn) {
		this.removeListener(this.CHANGE_EVENT, fn);
	}
});
