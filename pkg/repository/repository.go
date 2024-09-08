package repository

import (
	goproj "github.com/gfifgfifofich/GoProj"
	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user goproj.User) (string, error)
	GetUserRTokensByGUID(guid string) ([]string, error)
	UpdateUserRefreshTokens(guid string, rTokens []string) error
	GetUsersEmailByGUID(guid string) (string, error)
}

type Repository struct {
	Authorization
}

func NewRepository(pdatabase *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthDb(pdatabase),
	}
}
