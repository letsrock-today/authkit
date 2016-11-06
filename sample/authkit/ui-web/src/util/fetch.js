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
        if (!options) {
            options = {};
        }
        if (!options.credenitals) {
            options.credentials = 'same-origin';
        }
        if (!options.headers) {
            options.headers = {};
        }
		if (sameOrigin(url)) {
			let c = cookies.parse(document.cookie);
			options.headers[csrfHeaderName] = c[csrfCookieName];
			if (authTokenKey) {
				options.headers["Authorization"] =
					"Bearer " + localStorage.getItem(authTokenKey);
			}
		} else {
            console.warn('CSRF & Authorization headers skipped in cross-origin request');
		}
        return fetch(url, options);
    },

    setCSRFHeaderName: (name) => {
        headerName = name;
    },

    setCSRFCookieName: (name) => {
        csrfCookieName = name;
    },

    setAuthTokenKey: (name) => {
        authTokenKey = name;
    },

};

let csrfHeaderName = "X-CSRF-Token",
    csrfCookieName = "_csrf",
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
