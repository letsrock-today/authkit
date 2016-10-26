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
        let generalHandler = err => {
            let msg = 'HTTP error: ' +
                response.statusText + '.';
            return Promise.reject({
                error: e,
                msg: msg,
                cause: new Error(response.statusText)
            });
        }
        if (response.headers.get('Content-Type').startsWith('application/json')) {
            return Promise.resolve(response.json())
                .then(
                    err => {
                        return Promise.reject({
                            error: e,
                            msg: msgForCode(err.code, err.message),
                            cause: new Error(response.statusText)
                        });
                },
                generalHandler);
        }
        return Promise.resolve(response.text())
            .then(
                err => {
                    let msg = 'HTTP error: ' +
                        response.statusText + ', ' +
                        err + '.';
                    return Promise.reject({
                        error: e,
                        msg: msg,
                        cause: new Error(response.statusText)
                    });
            },
            generalHandler);
    }
}

// Sample function for error message customization.
// Error code could be mapped to localized message using some localization lib
// (like i18next). Here, for simplicity, we just provide more readable
// English text.
function msgForCode(code, fallback) {
    switch (code) {
        case 'invalid_req_param':
            //TODO: use subcodes for concrete rules?
            return 'Invalid request parameter.';
        case 'account_disabled':
            return 'Account created but not activated yet. ' +
                'Please follow the link sent you by email to activate account ' +
                'and than try to login again.';
        case 'duplicate_account':
            return 'Sorry, this username already in use.';
        case 'auth_err':
            return 'Incorrect combination of username and password.';
    }
    return 'Server reported unrecognized error: ' + fallback;
}
