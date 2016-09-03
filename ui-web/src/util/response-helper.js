'use strict';

module.exports = {
    // Helper method to check response status in fetch.
    handleStatus: (response) => {
        if (response.status === 200 || response.status === 0) {
            return Promise.resolve(response.json());
        } else {
            return Promise.reject({
                error: 'ERROR',
                cause: new Error(response.statusText)
            });
        }
    }
}
