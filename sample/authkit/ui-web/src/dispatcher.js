'use strict';

import Riot from 'riot';

let dispatcher = {
    TRY_LOGIN: 'TRY_LOGIN',
    SHOW_LOGIN_DIALOG: 'SHOW_LOGIN_DIALOG',
    DO_LOGOUT: 'DO_LOGOUT',
    PROFILE_CHANGED: 'PROFILE_CHANGED',
    SAVE_PROFILE: 'SAVE_PROFILE',
    GET_FRIENDS: 'GET_FRIENDS',
};

Riot.observable(dispatcher);

module.exports = dispatcher;
