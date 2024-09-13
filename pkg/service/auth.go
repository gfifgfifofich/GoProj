package service

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"

	goproj "github.com/gfifgfifofich/GoProj"
	"github.com/gfifgfifofich/GoProj/pkg/repository"
)

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (authService *AuthService) CreateUser(user goproj.User) (string, error) {
	user.Password = authService.generatePasswordHash(user.Password)
	return authService.repo.CreateUser(user)
}

func (AuthService *AuthService) generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func (authService *AuthService) GetUserRTokensByGUID(guid string) ([]string, error) {
	return authService.repo.GetUserRTokensByGUID(guid)
}
func (authService *AuthService) UpdateUserRefreshTokens(guid string, rTokens []string) error {
	return authService.repo.UpdateUserRefreshTokens(guid, rTokens)
}

// guid 32						 32	id/link 28				60 time 12   72 , the bcrypts maximum
// db033ddc20214537a63686fb3bcdeab0	db033ddc20214537a63686fb3b tttttttttttt
func CreateRefreshTokenFromData(guid string, id string, t int64) []byte {
	rtDataBaseString := make([]byte, 72) // bcrypt cant use more
	i := 0
	byteIter := 0
	// write guid without the '-'
	for byteIter < 32 {
		if guid[i] != '-' {
			rtDataBaseString[byteIter] = guid[i]
			byteIter++
		}
		i++
	}
	// put the link string in
	i = 0
	for i < 28 {
		rtDataBaseString[byteIter] = id[i]
		byteIter++
		i++
	}
	// put the Expiration date in
	// base64 to avoid random symbols that can interfere with sql
	// probably not a problem beause bcrypt changes stuff in db, but having special sumbols in string doesnt feel right
	ExpirationInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(ExpirationInBytes, uint64(t))
	ExpirationInBytesb64 := base64.StdEncoding.EncodeToString(ExpirationInBytes)
	i = 0
	for i < 12 {
		rtDataBaseString[byteIter] = ExpirationInBytesb64[i]
		byteIter++
		i++
	}
	return rtDataBaseString
}
