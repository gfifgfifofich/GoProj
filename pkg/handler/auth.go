package handler

import (
	"fmt"
	"log"
	"net/http"

	goproj "github.com/gfifgfifofich/GoProj"
	"github.com/gin-gonic/gin"
)

func (handler *Handler) signUp(ctx *gin.Context) {
	var input goproj.User
	if err := ctx.BindJSON(&input); err != nil {
		Response(ctx, http.StatusBadRequest, "invalid input body")
		log.Printf("error binging json: %s", err.Error())
		return
	}
	if handler == nil {
		log.Fatal("pHandler is nil")
	}
	if handler.service == nil {
		log.Fatal("pHandler.pservice is nil")
	}
	if handler.service.Authorization == nil {
		log.Fatal("pHandler.pservice.Authorization is nil")
	}
	id, err := handler.service.Authorization.CreateUser(input)
	if err != nil {
		Response(ctx, http.StatusInternalServerError, err.Error())
		log.Fatalf("internal error: %s", err.Error())
	}
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

/*
check query, parse url 							handler

	give user new refresh/access tokens			service
		save new refresh token in database		repository

write coockies									handler
communicate back to user about state of request handler
}
*/
func (handler *Handler) access(ctx *gin.Context) {
	qguid, ok := ctx.Request.URL.Query()["guid"]
	if !ok || len(qguid[0]) < 1 {
		Response(ctx, http.StatusBadRequest, "invalid input")
		return
	}

	var guid string = qguid[0]

	AccessTokenSigned, RefreshTokenSigned, AccessExpiration, RefreshExpiration, err := handler.service.Access(guid, ctx.ClientIP())
	if err != nil {
		Response(ctx, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "AccessToken",
		Value:    AccessTokenSigned,
		Expires:  AccessExpiration,
		HttpOnly: true,
		Secure:   true,
	})
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "RefreshToken",
		Value:    RefreshTokenSigned,
		Expires:  RefreshExpiration,
		HttpOnly: true,
		Secure:   true,
	})

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
	})

}

/*
get coockies									handler

	check tokens validity						service
		check refresh roken with database		repository
	update access token							service

save coockies									handler
}
*/
func (handler *Handler) refresh(ctx *gin.Context) {

	RefreshToken, err := ctx.Request.Cookie("RefreshToken")
	if err != nil {
		Response(ctx, http.StatusUnauthorized, "call Access route first")
		return
	}
	AccessToken, err := ctx.Request.Cookie("AccessToken")
	if err != nil {
		Response(ctx, http.StatusUnauthorized, "call Access route first")
		return
	}

	AcessTokenSigned, AccessTokenExpiration, err := handler.service.Refresh(RefreshToken.Value, AccessToken.Value, ctx.ClientIP())
	if err != nil {
		Response(ctx, http.StatusUnauthorized, fmt.Sprintf("Unauthorized: %s", err.Error()))
		return
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "AccessToken",
		Value:    AcessTokenSigned,
		Expires:  AccessTokenExpiration,
		HttpOnly: true,
		Secure:   true,
	})
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}
