'use strict';

import dispatcher from '../dispatcher';
import _fetch from '../util/fetch';
import respHelper from '../util/response-helper';
import cookies from 'cookie';

dispatcher.on(dispatcher.TRY_LOGIN, _login);
dispatcher.on(dispatcher.DO_LOGOUT, _logout);
dispatcher.on(dispatcher.DO_LOGIN, _doLogin);
dispatcher.on(dispatcher.SAVE_PROFILE, _saveProfile);
dispatcher.on(dispatcher.GET_FRIENDS, _getFriends);
dispatcher.on(dispatcher.CONFIRM_EMAIL, _confirmEmail);

////////////////////////////////////////////////////

function _saveProfile(profileForm, cb) {
    let body = new FormData(profileForm);
    _fetch.fetch('/api/profile', {
            method: 'POST',
            body: body
        })
        .then(r => {
            return respHelper.handleStatus(r);
        })
        .then(data => {
            data.authorized = true;
            dispatcher.trigger(dispatcher.PROFILE_CHANGED, data);
        })
        .catch(e => {
            console.log(e);
            if (e.error === 'AUTH') {
                alert('Auth error. Try to re-login');
                _logout();
            } else {
                cb('Error saving profile: ' + e.msg);
            }
        });
}

function _getFriends(cb) {
    _fetch.fetch('/api/friends')
        .then(r => {
            return respHelper.handleStatus(r);
        })
        .then(data => {
            cb(data, '');
        })
        .catch(e => {
            console.log(e);
            cb(null, 'Error fetching friends: ' + e.msg);
        });
}

function _confirmEmail(cb) {
    _fetch.fetch('/api/confirm-email', {
            method: 'POST'
        })
        .then(r => {
            return respHelper.handleStatus(r);
        })
        .then(data => {
            cb(data, '');
        })
        .catch(e => {
            console.log(e);
            cb(null, 'Error confirming email: ' + e.msg);
        });
}

function _doLogin(action, loginForm, cb) {
    let body = new FormData(loginForm);
    body.append('action', action);
    _fetch.fetch('/api/login-priv', {
        method: 'POST',
        body: body
    })
    .then(r => {
        return respHelper.handleStatus(r);
    })
    .then(data => {
        const {redirUrl} = data;
        window.location.href = redirUrl;
    })
    .catch(e => {
        console.log(e);
        cb(e.msg)
    });
}

function _login() {
    Promise.all([
        // See https://hacks.mozilla.org/2016/03/referrer-and-cache-control-apis-for-fetch/
        _fetch.fetch('/api/auth-providers'),
        _fetch.fetch('/api/auth-code-urls', { cache: "no-cache" })])
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

const authTokenKey = 'X-App-Auth-Token';

function _logout() {
    _fetch.fetch('/api/logout');
    localStorage.removeItem(authTokenKey);
    dispatcher.trigger(
        dispatcher.PROFILE_CHANGED, {
            username: '',
            authorized: false
        });
}

(window.onpopstate = function () {
    const authCookieName = 'X-App-Auth';
    _fetch.setAuthTokenKey(authTokenKey);
    let token = cookies.parse(document.cookie)[authCookieName];
    if (token) {
        localStorage.setItem(authTokenKey, token);
        document.cookie = cookies.serialize(authCookieName, "", {expires: new Date(0)});
    } else {
        token = localStorage.getItem(authTokenKey);
    }
    if (token) {
        _fetch.fetch('/api/profile')
            .then(r => {
                return respHelper.handleStatus(r);
            })
            .then(data => {
                data.authorized = true;
                dispatcher.trigger(dispatcher.PROFILE_CHANGED, data);
            })
            .catch(e => {
                console.log(e);
                if (e.error === 'AUTH') {
                    alert('Auth error. Try to re-login');
                    _logout();
                }
            });
    }
})();
