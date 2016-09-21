'use strict';

import Riot from 'riot';

let dispatcher = {
    TRY_LOGIN: 'TRY_LOGIN',
    SHOW_LOGIN_DIALOG: 'SHOW_LOGIN_DIALOG',
    DO_LOGOUT: 'DO_LOGOUT',
    USER_DATA_CHANGED: 'USER_DATA_CHANGED',
};

Riot.observable(dispatcher);

module.exports = dispatcher;
