package handler

import (
	"fmt"
	"log"
	"net/http"

	goproj "github.com/gfifgfifofich/GoProj"
	"github.com/gin-gonic/gin"
)

func (pHandler *Handler) signUp(pCtx *gin.Context) {
	var input goproj.User
	if err := pCtx.BindJSON(&input); err != nil {
		Response(pCtx, http.StatusBadRequest, "invalid input body")
		log.Printf("error binging json: %s", err.Error())
		return
	}
	if pHandler == nil {
		log.Fatal("pHandler is nil")
	}
	if pHandler.pservice == nil {
		log.Fatal("pHandler.pservice is nil")
	}
	if pHandler.pservice.Authorization == nil {
		log.Fatal("pHandler.pservice.Authorization is nil")
	}
	id, err := pHandler.pservice.Authorization.CreateUser(input)
	if err != nil {
		Response(pCtx, http.StatusInternalServerError, err.Error())
		log.Fatalf("internal error: %s", err.Error())
	}
	pCtx.JSON(http.StatusOK, map[string]interface{}{
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
func (pHandler *Handler) access(pCtx *gin.Context) {
	qguid, ok := pCtx.Request.URL.Query()["guid"]
	if !ok || len(qguid[0]) < 1 {
		Response(pCtx, http.StatusBadRequest, "invalid input")
		return
	}

	var guid string = qguid[0]

	atSigned, rtSigned, atExpiration, rtExpiration, err := pHandler.pservice.Access(guid, pCtx.ClientIP())
	if err != nil {
		Response(pCtx, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	http.SetCookie(pCtx.Writer, &http.Cookie{
		Name:     "tAccess",
		Value:    atSigned,
		Expires:  atExpiration,
		HttpOnly: true,
	})
	http.SetCookie(pCtx.Writer, &http.Cookie{
		Name:     "tRefresh",
		Value:    rtSigned,
		Expires:  rtExpiration,
		HttpOnly: true,
	})

	pCtx.JSON(http.StatusOK, map[string]interface{}{
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
func (pHandler *Handler) refresh(pCtx *gin.Context) {

	tRefresh, err := pCtx.Request.Cookie("tRefresh")
	if err != nil {
		Response(pCtx, http.StatusUnauthorized, "Unauthorized1")
		return
	}
	tAccess, err := pCtx.Request.Cookie("tAccess")
	if err != nil {
		Response(pCtx, http.StatusUnauthorized, fmt.Sprintf("Unauthorized2: %s", err.Error()))
		return
	}

	atSigned, _, atExpiration, _, err := pHandler.pservice.Refresh(tRefresh.Value, tAccess.Value, pCtx.ClientIP())
	if err != nil {
		Response(pCtx, http.StatusUnauthorized, "Unauthorized3")
		return
	}

	http.SetCookie(pCtx.Writer, &http.Cookie{
		Name:     "tAccess",
		Value:    atSigned,
		Expires:  atExpiration,
		HttpOnly: true,
	})

	pCtx.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}
