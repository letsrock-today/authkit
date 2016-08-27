'use strict';

import dispatcher from '../dispatcher';

dispatcher.on(dispatcher.TRY_LOGIN, _login);
dispatcher.on(dispatcher.DO_LOGOUT, _logout);

////////////////////////////////////////////////////

function _login() {
    let p = [{
            "id": "google",
            "name": "Google+ (https://plus.google.com)",
            "iconUrl": "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz48IURPQ1RZUEUgc3ZnIFBVQkxJQyAiLS8vVzNDLy9EVEQgU1ZHIDEuMS8vRU4iICJodHRwOi8vd3d3LnczLm9yZy9HcmFwaGljcy9TVkcvMS4xL0RURC9zdmcxMS5kdGQiPjxzdmcgdmVyc2lvbj0iMS4xIiBpZD0iTGF5ZXJfMSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxuczp4bGluaz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94bGluayIgeD0iMHB4IiB5PSIwcHgiIHdpZHRoPSIxMzQuNjU4cHgiIGhlaWdodD0iMTMxLjY0NnB4IiB2aWV3Qm94PSIwIDAgMTM0LjY1OCAxMzEuNjQ2IiBlbmFibGUtYmFja2dyb3VuZD0ibmV3IDAgMCAxMzQuNjU4IDEzMS42NDYiIHhtbDpzcGFjZT0icHJlc2VydmUiPjxnPjxwYXRoIGZpbGw9IiNEQzRBMzgiIGQ9Ik0xMjYuNTE1LDQuMTA5SDguMTQ0Yy0yLjE3NywwLTMuOTQsMS43NjMtMy45NCwzLjkzOHYxMTUuNTQ2YzAsMi4xNzksMS43NjMsMy45NDIsMy45NCwzLjk0MmgxMTguMzcxYzIuMTc3LDAsMy45NC0xLjc2NCwzLjk0LTMuOTQyVjguMDQ4QzEzMC40NTUsNS44NzIsMTI4LjY5MSw0LjEwOSwxMjYuNTE1LDQuMTA5eiIvPjxnPjxwYXRoIGZpbGw9IiNGRkZGRkYiIGQ9Ik03MC40NzksNzEuODQ1bC0zLjk4My0zLjA5M2MtMS4yMTMtMS4wMDYtMi44NzItMi4zMzQtMi44NzItNC43NjVjMC0yLjQ0MSwxLjY1OS0zLjk5MywzLjA5OS01LjQzYzQuNjQtMy42NTIsOS4yNzYtNy41MzksOS4yNzYtMTUuNzNjMC04LjQyMy01LjMtMTIuODU0LTcuODQtMTQuOTU2aDYuODQ5bDcuMTg5LTQuNTE3SDYwLjQxOGMtNS45NzYsMC0xNC41ODgsMS40MTQtMjAuODkzLDYuNjE5Yy00Ljc1Miw0LjEtNy4wNyw5Ljc1My03LjA3LDE0Ljg0MmMwLDguNjM5LDYuNjMzLDE3LjM5NiwxOC4zNDYsMTcuMzk2YzEuMTA2LDAsMi4zMTYtMC4xMDksMy41MzQtMC4yMjJjLTAuNTQ3LDEuMzMxLTEuMSwyLjQzOS0xLjEsNC4zMmMwLDMuNDMxLDEuNzYzLDUuNTM1LDMuMzE3LDcuNTI4Yy00Ljk3NywwLjM0Mi0xNC4yNjgsMC44OTMtMjEuMTE3LDUuMTAzYy02LjUyMywzLjg3OS04LjUwOCw5LjUyNS04LjUwOCwxMy41MWMwLDguMjAyLDcuNzMxLDE1Ljg0MiwyMy43NjIsMTUuODQyYzE5LjAxLDAsMjkuMDc0LTEwLjUxOSwyOS4wNzQtMjAuOTMyQzc5Ljc2NCw3OS43MDksNzUuMzQ0LDc1Ljk0Myw3MC40NzksNzEuODQ1eiBNNTYsNTkuMTA3Yy05LjUxLDAtMTMuODE4LTEyLjI5NC0xMy44MTgtMTkuNzEyYzAtMi44ODgsMC41NDctNS44NywyLjQyOC04LjE5OWMxLjc3My0yLjIxOCw0Ljg2MS0zLjY1Nyw3Ljc0NC0zLjY1N2M5LjE2OCwwLDEzLjkyMywxMi40MDQsMTMuOTIzLDIwLjM4MmMwLDEuOTk2LTAuMjIsNS41MzMtMi43NjIsOC4wOUM2MS43MzcsNTcuNzg1LDU4Ljc2Miw1OS4xMDcsNTYsNTkuMTA3eiBNNTYuMTA5LDEwMy42NWMtMTEuODI2LDAtMTkuNDUyLTUuNjU3LTE5LjQ1Mi0xMy41MjNjMC03Ljg2NCw3LjA3MS0xMC41MjQsOS41MDQtMTEuNDA1YzQuNjQtMS41NjEsMTAuNjExLTEuNzc5LDExLjYwNy0xLjc3OWMxLjEwNSwwLDEuNjU4LDAsMi41MzgsMC4xMTFjOC40MDcsNS45ODMsMTIuMDU2LDguOTY1LDEyLjA1NiwxNC42MjlDNzIuMzYyLDk4LjU0Miw2Ni43MjMsMTAzLjY1LDU2LjEwOSwxMDMuNjV6Ii8+PHBvbHlnb24gZmlsbD0iI0ZGRkZGRiIgcG9pbnRzPSI5OC4zOTMsNTguOTM4IDk4LjM5Myw0Ny44NjMgOTIuOTIzLDQ3Ljg2MyA5Mi45MjMsNTguOTM4IDgxLjg2Niw1OC45MzggODEuODY2LDY0LjQ2OSA5Mi45MjMsNjQuNDY5IDkyLjkyMyw3NS42MTIgOTguMzkzLDc1LjYxMiA5OC4zOTMsNjQuNDY5IDEwOS41MDYsNjQuNDY5IDEwOS41MDYsNTguOTM4ICIvPjwvZz48L2c+PC9zdmc+"
                }, {
            "id": "fb",
            "name": "FB (https://www.facebook.com/)",
            "iconUrl": "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiA/PjxzdmcgaGVpZ2h0PSI1MTIiIGlkPSJMYXllcl8xIiB2ZXJzaW9uPSIxLjEiIHZpZXdCb3g9IjAgMCA1MTIgNTEyIiB3aWR0aD0iNTEyIiB4bWw6c3BhY2U9InByZXNlcnZlIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOmNjPSJodHRwOi8vY3JlYXRpdmVjb21tb25zLm9yZy9ucyMiIHhtbG5zOmRjPSJodHRwOi8vcHVybC5vcmcvZGMvZWxlbWVudHMvMS4xLyIgeG1sbnM6aW5rc2NhcGU9Imh0dHA6Ly93d3cuaW5rc2NhcGUub3JnL25hbWVzcGFjZXMvaW5rc2NhcGUiIHhtbG5zOnJkZj0iaHR0cDovL3d3dy53My5vcmcvMTk5OS8wMi8yMi1yZGYtc3ludGF4LW5zIyIgeG1sbnM6c29kaXBvZGk9Imh0dHA6Ly9zb2RpcG9kaS5zb3VyY2Vmb3JnZS5uZXQvRFREL3NvZGlwb2RpLTAuZHRkIiB4bWxuczpzdmc9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZGVmcyBpZD0iZGVmczEyIi8+PGcgaWQ9Imc1OTkxIj48cmVjdCBoZWlnaHQ9IjUxMiIgaWQ9InJlY3QyOTg3IiByeD0iNjQiIHJ5PSI2NCIgc3R5bGU9ImZpbGw6IzNiNTk5ODtmaWxsLW9wYWNpdHk6MTtmaWxsLXJ1bGU6bm9uemVybztzdHJva2U6bm9uZSIgd2lkdGg9IjUxMiIgeD0iMCIgeT0iMCIvPjxwYXRoIGQ9Ik0gMjg2Ljk2NzgzLDQ1NS45OTk3MiBWIDI3My41Mzc1MyBoIDYxLjI0NCBsIDkuMTY5OSwtNzEuMTAyNjYgaCAtNzAuNDEyNDYgdiAtNDUuMzk0OTMgYyAwLC0yMC41ODgyOCA1LjcyMDY2LC0zNC42MTk0MiAzNS4yMzQ5NiwtMzQuNjE5NDIgbCAzNy42NTU0LC0wLjAxMTIgViA1OC44MDc5MTUgYyAtNi41MDk3LC0wLjg3MzgxIC0yOC44NTcxLC0yLjgwNzk0IC01NC44Njc1LC0yLjgwNzk0IC01NC4yODgwMywwIC05MS40NDk5NSwzMy4xNDU4NSAtOTEuNDQ5OTUsOTMuOTk4MTI1IHYgNTIuNDM3MDggaCAtNjEuNDAxODEgdiA3MS4xMDI2NiBoIDYxLjQwMDM5IHYgMTgyLjQ2MjE5IGggNzMuNDI3MDcgeiIgaWQ9ImZfMV8iIHN0eWxlPSJmaWxsOiNmZmZmZmYiLz48L2c+PC9zdmc+"
                }, {
            "id": "linkedin",
            "name": "LN (https://www.linkedin.com/)",
            "iconUrl": "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiA/PjxzdmcgaGVpZ2h0PSI1MTIiIGlkPSJMYXllcl8xIiB2ZXJzaW9uPSIxLjEiIHZpZXdCb3g9IjAgMCA1MTIgNTEyIiB3aWR0aD0iNTEyIiB4bWw6c3BhY2U9InByZXNlcnZlIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOmNjPSJodHRwOi8vY3JlYXRpdmVjb21tb25zLm9yZy9ucyMiIHhtbG5zOmRjPSJodHRwOi8vcHVybC5vcmcvZGMvZWxlbWVudHMvMS4xLyIgeG1sbnM6aW5rc2NhcGU9Imh0dHA6Ly93d3cuaW5rc2NhcGUub3JnL25hbWVzcGFjZXMvaW5rc2NhcGUiIHhtbG5zOnJkZj0iaHR0cDovL3d3dy53My5vcmcvMTk5OS8wMi8yMi1yZGYtc3ludGF4LW5zIyIgeG1sbnM6c29kaXBvZGk9Imh0dHA6Ly9zb2RpcG9kaS5zb3VyY2Vmb3JnZS5uZXQvRFREL3NvZGlwb2RpLTAuZHRkIiB4bWxuczpzdmc9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZGVmcyBpZD0iZGVmczEyIi8+PGcgaWQ9Imc1ODkxIj48cmVjdCBoZWlnaHQ9IjUxMiIgaWQ9InJlY3QyOTg3IiByeD0iNjQiIHJ5PSI2NCIgc3R5bGU9ImZpbGw6IzAwODNiZTtmaWxsLW9wYWNpdHk6MTtmaWxsLXJ1bGU6bm9uemVybztzdHJva2U6bm9uZSIgd2lkdGg9IjUxMiIgeD0iMCIgeT0iNS42ODQzNDE5ZS0wMTQiLz48ZyBpZD0iZzktMSIgdHJhbnNmb3JtPSJtYXRyaXgoMS41NTM3OTQ2LDAsMCwxLjU1Mzc5NDYsLTE0MC44NzMzMiwtMTMyLjY0NTUyKSI+PHJlY3QgaGVpZ2h0PSIxNjYuMDIxIiBpZD0icmVjdDExIiBzdHlsZT0iZmlsbDojZmZmZmZmIiB3aWR0aD0iNTUuMTk0IiB4PSIxMjkuOTU3IiB5PSIyMDAuMzU2OTkiLz48cGF0aCBkPSJtIDE1Ny45MjcsMTIwLjMwMyBjIC0xOC44ODQsMCAtMzEuMjIyLDEyLjQxNSAtMzEuMjIyLDI4LjY4NyAwLDE1LjkzIDExLjk2MywyOC42ODcgMzAuNDkxLDI4LjY4NyBoIDAuMzU3IGMgMTkuMjQ1LDAgMzEuMjI0LC0xMi43NTcgMzEuMjI0LC0yOC42ODcgLTAuMzU3LC0xNi4yNzIgLTExLjk3OCwtMjguNjg3IC0zMC44NSwtMjguNjg3IHoiIGlkPSJwYXRoMTMtMCIgc3R5bGU9ImZpbGw6I2ZmZmZmZiIvPjxwYXRoIGQ9Im0gMzIwLjYwNCwxOTYuNDUzIGMgLTI5LjI3NywwIC00Mi4zOTEsMTYuMTAxIC00OS43MzQsMjcuNDEgdiAtMjMuNTA2IGggLTU1LjE4IGMgMC43MzIsMTUuNTczIDAsMTY2LjAyMSAwLDE2Ni4wMjEgaCA1NS4xNzkgViAyNzMuNjYgYyAwLC00Ljk2MyAwLjM1NywtOS45MjQgMS44MiwtMTMuNDcxIDMuOTgyLC05LjkxMSAxMy4wNjgsLTIwLjE3OCAyOC4zMTMsLTIwLjE3OCAxOS45NTksMCAyNy45NTUsMTUuMjMgMjcuOTU1LDM3LjUzOSB2IDg4LjgyOCBoIDU1LjE4MiB2IC05NS4yMDYgYyAwLC01MC45OTYgLTI3LjIyNywtNzQuNzE5IC02My41MzUsLTc0LjcxOSB6IiBpZD0icGF0aDE1IiBzdHlsZT0iZmlsbDojZmZmZmZmIi8+PC9nPjwvZz48L3N2Zz4="
                }],

        u = [{
            "id": "google",
            "url": "https://accounts.google.com/o/oauth2/auth?access_type=online\u0026client_id=465779685744-agagiio9hi7i7p7u9di4rffklge5pnq8.apps.googleusercontent.com\u0026redirect_uri=https%3A%2F%2Fletsrock.today%2Fauth%2Fv1%2Flogin\u0026response_type=code\u0026scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fplus.login\u0026state=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NzIzMDc2MDIsImlzcyI6ImxldHNyb2NrLnRvZGF5IiwibHJfY2xpZW50IjoiIiwicGlkIjoiZ29vZ2xlIiwic2FsdCI6IlhROFFjbFVKIn0.j-mQa69snYF5qmLzYWeLxV0APTWsaYfSV7AojiWip8g"
                }, {
            "id": "fb",
            "url": "https://www.facebook.com/dialog/oauth?access_type=online\u0026client_id=1453175985002027\u0026redirect_uri=https%3A%2F%2Fletsrock.today%2Fauth%2Fv1%2Flogin\u0026response_type=code\u0026state=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NzIzMDc2MDIsImlzcyI6ImxldHNyb2NrLnRvZGF5IiwibHJfY2xpZW50IjoiIiwicGlkIjoiZmIiLCJzYWx0IjoiSnZrZGh3NUpLa3lIZmsrbCJ9.4mGbWz-rFS58sWNnwhrKN-v3Im9gcRE1fYs6TTz9uWg"
                }, {
            "id": "linkedin",
            "url": "https://www.linkedin.com/uas/oauth2/authorization?access_type=online\u0026client_id=777vshr3cokh4g\u0026redirect_uri=https%3A%2F%2Fletsrock.today%2Fauth%2Fv1%2Flogin\u0026response_type=code\u0026state=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NzIzMDc2MDIsImlzcyI6ImxldHNyb2NrLnRvZGF5IiwibHJfY2xpZW50IjoiIiwicGlkIjoibGlua2VkaW4iLCJzYWx0IjoiR1N5b052SjVJM2E2RCsreW9nTT0ifQ.Xux5GGListehy5ceqk-Hmji0fPLWJe3CiH9zODONrkA"
                }];

    p.forEach((v) => {
        let it = u.find((e) => {
            return e.id === v.id ? e.url : null;
        });
        if (it) {
            v.authCodeUrl = it.url;
        }
    });

    dispatcher.trigger(dispatcher.SHOW_LOGIN_DIALOG, p);
}

function _logout() {
    dispatcher.trigger(
        dispatcher.USER_DATA_CHANGED, {
            username: '',
            authorized: false
        });
}
