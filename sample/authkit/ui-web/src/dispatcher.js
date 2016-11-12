'use strict';

import Riot from 'riot';

let dispatcher = {
    TRY_LOGIN: 'TRY_LOGIN',
    SHOW_LOGIN_DIALOG: 'SHOW_LOGIN_DIALOG',
    DO_LOGIN: 'DO_LOGIN',
    DO_LOGOUT: 'DO_LOGOUT',
    PROFILE_CHANGED: 'PROFILE_CHANGED',
    SAVE_PROFILE: 'SAVE_PROFILE',
    GET_FRIENDS: 'GET_FRIENDS',
    CONFIRM_EMAIL: 'CONFIRM_EMAIL',
};

Riot.observable(dispatcher);

module.exports = dispatcher;
