package route

import (
	"github.com/labstack/echo"

	"github.com/letsrock-today/hydra-sample/backend/service/user/userapi"
)

func Init(e *echo.Echo, ua userapi.UserAPI) {
	initMiddleware(e)
	initReverseProxy(e)
	initStatic(e)
	initAPI(e, ua)
}
