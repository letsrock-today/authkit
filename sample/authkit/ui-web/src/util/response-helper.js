'use strict';

module.exports = {
    // Helper method to check response status in fetch.
    handleStatus: (response) => {
        let s = response.status, e;
        if (s === 200 || s === 0) {
            return Promise.resolve(response.json());
        } else if (s === 401 || s === 403) {
            e = 'AUTH';
        } else {
            e = 'ERROR';
        }
        return Promise.resolve(response.json())
        .then(
            err => {
                // err.code can be used to provide i18ned message
                return Promise.reject({
                    error: e,
                    msg: err.message,
                    cause: new Error(response.statusText)
                });
            },
            err => {
                return Promise.reject({
                    error: e,
                    msg: response.statusText,
                    cause: new Error(response.statusText)
                });
            });
    }
}
