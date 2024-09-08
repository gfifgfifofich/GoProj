package handler

import (
	"github.com/gfifgfifofich/GoProj/pkg/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	pservice *service.Service
}

func NewHandler(pservice *service.Service) *Handler {
	return &Handler{pservice: pservice}
}

func (phandler *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", phandler.signUp)
	}
	task := router.Group("/task")
	{
		task.POST("/access", phandler.access)
		task.POST("/refresh", phandler.refresh)
	}
	return router
}
