'use strict';

import dispatcher from '../dispatcher';
import _fetch from '../util/fetch';
import respHelper from '../util/response-helper';
import cookies from 'cookie';

dispatcher.on(dispatcher.TRY_LOGIN, _login);
dispatcher.on(dispatcher.DO_LOGOUT, _logout);
dispatcher.on(dispatcher.DO_USERDATA_UPDATE, _save);

////////////////////////////////////////////////////

function _login() {
    Promise.all([
        // See https://hacks.mozilla.org/2016/03/referrer-and-cache-control-apis-for-fetch/
        _fetch.fetch('/api/auth-providers'),
        _fetch.fetch('/api/auth-code-urls', {
            cache: "no-cache"
        })])
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

const authCookieName = 'X-App-Auth';

function _logout() {
    document.cookie = cookies.serialize(authCookieName, "", {expires: new Date(0)});
    dispatcher.trigger(
        dispatcher.USER_DATA_CHANGED, {
            username: '',
            authorized: false
        });
}

function _save(profile) {
    _fetch.fetch('/api/profile', {
        method: 'POST',
        body: profile
    })
    .then(r => {
        return respHelper.handleStatus(r);
    })
    .then(data => {
        const {error} = data;
        if (error) {
            console.log(e);
        }
    })
    .catch(e => {
        console.log(e);
    });
}

(window.onpopstate = function () {
    _fetch.setAuthCookieName(authCookieName);
    let token = cookies.parse(document.cookie)[authCookieName];
    if (token) {
        _fetch.fetch('/api/profile')
        .then(r => {
            return respHelper.handleStatus(r);
        })
        .then(data => {
            const {error} = data;
            if (error) {
                console.log(error);
            } else {
                data.authorized = true;
                dispatcher.trigger(dispatcher.USER_DATA_CHANGED, data);
            }
        })
        .catch(e => {
            console.log(e);
        });
    }
})();
