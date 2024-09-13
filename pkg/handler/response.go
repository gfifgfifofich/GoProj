package handler

import "github.com/gin-gonic/gin"

func Response(Ctx *gin.Context, statusCode int, message string) {

	Ctx.AbortWithStatusJSON(statusCode, message)
}
