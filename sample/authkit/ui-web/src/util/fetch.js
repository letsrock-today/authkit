'use strict';

import 'whatwg-fetch';
import cookies from 'cookie';

module.exports = {

    // Custom fetch, which automatically sends CSRF and auth headers.
    // To be used with https://echo.labstack.com/middleware/csrf.
    // Note: headers are skipped in case of cross-origin requests to prevent
    // token leakage. Consider to use buildin or polyfill fetch for cross-origin
    // requests.
    fetch: (url, options) => {
		if (!sameOrigin(url)) {
            console.warn('DO NOT use this fetch implementation for cross-origin requests');
            return fetch(url, options);
        }
        if (!options) {
            options = {};
        }
        if (!options.credenitals) {
            options.credentials = 'same-origin';
        }
        if (!options.headers) {
            options.headers = new Headers();
        }
        options.headers = new Headers(options.headers);
		let c = cookies.parse(document.cookie);
		options.headers.set(csrfHeaderName, c[csrfCookieName]);
		if (authTokenKey) {
			options.headers.set("Authorization",
				"Bearer " + localStorage.getItem(authTokenKey));
		}
        return fetch(url, options).then(response => {
            let token = response.headers.get(authUpdateHeaderName);
            if (token) {
                localStorage.setItem(authTokenKey, token);
            }
            return response;
        });
    },

    setCSRFHeaderName: (name) => {
        headerName = name;
    },

    setCSRFCookieName: (name) => {
        csrfCookieName = name;
    },

    setAuthUpdateHeaderName: (name) => {
        authUpdateHeaderName = name;
    },

    setAuthTokenKey: (name) => {
        authTokenKey = name;
    },
};

let csrfHeaderName = "X-CSRF-Token",
    csrfCookieName = "_csrf",
    authUpdateHeaderName = "X-App-Auth-Update",
    authTokenKey,
    loc = window.location,
    a = document.createElement('a');

function sameOrigin(url) {
    a.href = url
    return a.hostname === loc.hostname &&
           a.port === loc.port &&
           a.protocol === loc.protocol &&
           loc.protocol !== 'file:'
}
