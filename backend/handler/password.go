package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/letsrock-today/hydra-sample/backend/config"
	"github.com/letsrock-today/hydra-sample/backend/util"
)

type (
	resetPasswordForm struct {
		Email []string `form:"email" valid:"required,email"`
	}
)

var (
	emailInvalidErr    = fmt.Errorf("Email invalid or not registered in the app")
	confirmPasswordURL = "/password-confirm"
)

func ResetPassword(c echo.Context) error {
	var rp resetPasswordForm
	if err := c.Bind(&rp); err != nil {
		return err
	}
	log.Printf("resetPasswordForm: %#v\n", rp)
	if _, err := govalidator.ValidateStruct(rp); err != nil {
		return c.JSON(http.StatusOK, newJsonError(err))
	}

	email := rp.Email[0]
	user, err := UserService.GetUser(email)
	if err != nil {
		return err
	}
	if user.PasswordHash == "" {
		return c.JSON(http.StatusOK, newJsonError(emailInvalidErr))
	}
	cfg := config.GetConfig()
	confirmPasswordExternalURL := cfg.ExternalBaseURL + confirmPasswordURL
	link := fmt.Sprintf("%s?email=%s&hash=%s", confirmPasswordExternalURL, email, user.PasswordHash)
	text := fmt.Sprintf("Follow this link to change your password: %s\n", link)
	if err = util.SendEmail(email, "Confirm password reset", text); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, struct{}{})
}
