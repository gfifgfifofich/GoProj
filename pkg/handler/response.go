package handler

import "github.com/gin-gonic/gin"

func Response(pCtx *gin.Context, statusCode int, message string) {

	pCtx.AbortWithStatusJSON(statusCode, message)
}
