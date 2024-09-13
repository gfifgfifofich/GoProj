package handler

import (
	"github.com/gfifgfifofich/GoProj/pkg/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", handler.signUp)
	}
	task := router.Group("/task")
	{
		task.POST("/access", handler.access)
		task.POST("/refresh", handler.refresh)
	}
	return router
}
