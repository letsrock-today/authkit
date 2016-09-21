'use strict';

import 'whatwg-fetch';

module.exports = {

    // custom fetch, which automatically sends CSRF header
    // to be used with https://echo.labstack.com/middleware/csrf
    fetch: (url, options) => {
        if (!options) {
            options = {};
        }
        if (!options.headers) {
            options.headers = {};
        }
        options.headers[csrfHeaderName] = getCookie(csrfCookieName);
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
    }

};

function getCookie(name) {
    function escape(s) {
        return s.replace(/([.*+?\^${}()|\[\]\/\\])/g, '\\$1');
    };
    let match = document.cookie.match(RegExp('(?:^|;\\s*)' + escape(name) + '=([^;]*)'));
    return match ? match[1] : null;
}

let csrfHeaderName = "X-CSRF-Token",
    csrfCookieName = "csrf";
