'use strict';

module.exports = {
    // Helper method to check response status in fetch.
    handleStatus: (response) => {
        if (response.status === 200 || response.status === 0) {
            return Promise.resolve(response.json());
        } else if (response.status === 403) {
            return Promise.reject({
                error: 'AUTH',
                msg: response.statusText,
                cause: new Error(response.statusText)
            });
        } else {
            return Promise.reject({
                error: 'ERROR',
                msg: response.statusText,
                cause: new Error(response.statusText)
            });
        }
    }
}
