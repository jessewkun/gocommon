package response

import (
	"github.com/gin-gonic/gin"
)

// BindAndValidate 统一绑定和校验（自动判断来源类型）
func BindAndValidate(ctx *gin.Context, obj interface{}) error {
	var err error

	switch ctx.Request.Method {
	case "GET":
		err = ctx.ShouldBindQuery(obj)
	case "POST", "PUT", "PATCH":
		err = ctx.ShouldBindJSON(obj)
	default:
		err = ctx.ShouldBind(obj) // 支持 form-data 等
	}

	return err
}
