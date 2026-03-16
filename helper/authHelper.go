package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckUserType(ctx *gin.Context, role string) (err error) {
	// Get user role from context
	userType := ctx.GetString("user_type") // This retrieves "user_type" from the request context.
	err = nil

	if userType != role {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	return err
}

func MatchUserTypeToUid(ctx *gin.Context, userId string) (err error) {

	userType := ctx.GetString("user_type")
	uid := ctx.GetString("uid")
	err = nil

	// the below condition is: he is normal user and trying to access others info condition true.
	if userType == "USER" && uid != userId {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	// err = CheckUserType(ctx, userType)
	return err
}
