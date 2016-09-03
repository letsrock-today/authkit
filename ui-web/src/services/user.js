'use strict';

import dispatcher from '../dispatcher';
import 'whatwg-fetch';
import respHelper from '../util/response-helper';

dispatcher.on(dispatcher.TRY_LOGIN, _login);
dispatcher.on(dispatcher.DO_LOGOUT, _logout);

////////////////////////////////////////////////////

function _login() {
    Promise.all([
            // See https://hacks.mozilla.org/2016/03/referrer-and-cache-control-apis-for-fetch/
            fetch('/api/auth-providers'),
            fetch('/api/auth-code-urls', {cache: "no-cache"})])
        .then(responses => {
            return Promise.all([
                    respHelper.handleStatus(responses[0]),
                    respHelper.handleStatus(responses[1])]);
        })
        .then(data => {
            let obj1 = data[0],
                obj2 = data[1];
            let p = obj1.providers,
                u = obj2.urls;
            p.forEach((v) => {
                let it = u.find((e) => {
                    return e.id === v.id ? e.url : null;
                });
                if (it) {
                    v.authCodeUrl = it.url;
                }
            });
            dispatcher.trigger(dispatcher.SHOW_LOGIN_DIALOG, p);
        })
        .catch(e => {
            console.log(e);
        });
}

function _logout() {
    dispatcher.trigger(
        dispatcher.USER_DATA_CHANGED, {
            username: '',
            authorized: false
        });
}
