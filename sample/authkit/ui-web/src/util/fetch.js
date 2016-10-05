'use strict';

import 'whatwg-fetch';
import cookies from 'cookie';

module.exports = {

    // custom fetch, which automatically sends CSRF and auth headers
    // to be used with https://echo.labstack.com/middleware/csrf
    fetch: (url, options) => {
        if (!options) {
            options = {};
        }
        if (!options.headers) {
            options.headers = {};
        }
        let c = cookies.parse(document.cookie);
        options.headers[csrfHeaderName] = c[csrfCookieName];
        if (authCookieName) {
            options.headers["Authorization"] = "Bearer " + c[authCookieName];
        }
        if (!options.credenitals) {
            options.credentials = 'same-origin';
        }
        return fetch(url, options);
    },

    setCSRFHeaderName: (name) => {
        headerName = name;
    },

    setCSRFCookieName: (name) => {
        csrfCookieName = name;
    },

    setAuthCookieName: (name) => {
        authCookieName = name;
    },

};

let csrfHeaderName = "X-CSRF-Token",
    csrfCookieName = "csrf",
    authCookieName;
