package service

import (
	"time"

	goproj "github.com/gfifgfifofich/GoProj"
	"github.com/gfifgfifofich/GoProj/pkg/repository"
)

type Authorization interface {
	CreateUser(user goproj.User) (string, error)
	Access(guid string, clientIP string) (string, string, time.Time, time.Time, error)
	Refresh(usrRToken string, aToken string, clientIP string) (string, string, time.Time, time.Time, error)
}

type Service struct {
	Authorization
}

func NewService(prepo *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(prepo.Authorization),
	}
}
