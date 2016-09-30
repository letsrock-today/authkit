'use strict';

import Riot from 'riot';

let dispatcher = {
    TRY_LOGIN: 'TRY_LOGIN',
    SHOW_LOGIN_DIALOG: 'SHOW_LOGIN_DIALOG',
    DO_LOGOUT: 'DO_LOGOUT',
    USER_DATA_CHANGED: 'USER_DATA_CHANGED',
    DO_USERDATA_UPDATE: 'DO_USERDATA_UPDATE',
};

Riot.observable(dispatcher);

module.exports = dispatcher;
