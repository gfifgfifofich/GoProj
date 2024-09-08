package service

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/dgrijalva/jwt-go"
	goproj "github.com/gfifgfifofich/GoProj"
	"github.com/gfifgfifofich/GoProj/pkg/repository"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (pauthService *AuthService) CreateUser(user goproj.User) (int, error) {
	user.Password = pauthService.generatePasswordHash(user.Password)

	return pauthService.repo.CreateUser(user)
}
func (pAuthService *AuthService) generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func (pauthService *AuthService) GetUserRTokensByGUID(guid string) ([]string, error) {
	return pauthService.repo.GetUserRTokensByGUID(guid)
}
func (pauthService *AuthService) UpdateUserRefreshTokens(guid string, rTokens []string) error {
	return pauthService.repo.UpdateUserRefreshTokens(guid, rTokens)
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

func (pauthService *AuthService) Access(guid string, clientIP string) (string, string, time.Time, time.Time, error) {

	rtExpiration := time.Now().Add(5 * time.Hour)

	linkBytes := make([]byte, 28)
	_, err := rand.Read(linkBytes)
	if err != nil {
		return "", "", time.Now(), time.Now(), err
	}
	// token id generation based on uuid/guid
	NewUnprocessedGuid, err := uuid.NewV4()
	if err != nil {
		log.Print("Failed to generate guid at service/access")
	}
	NewUnprocessedByteGuid := NewUnprocessedGuid.String()
	// get new id for tokeni
	i := 0
	byteIter := 0
	for byteIter < 28 {
		if NewUnprocessedByteGuid[i] != '-' {
			linkBytes[byteIter] = NewUnprocessedByteGuid[i]
			byteIter++
		}
		i++
	}

	rtClaims := &сlaims{
		UserID: guid,
		UserIP: clientIP,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: rtExpiration.Unix(),
			Id:        string(linkBytes),
		},
	}

	//user will have a jwt token, DB will have similar token, not jwt, just bcrypt
	//jwt is base64 and cant be changed

	// user side refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, rtClaims)
	rtSigned, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		log.Printf("refresh token error: %v\n", err)
		return "", "", time.Now(), time.Now(), err
	}

	atExpiration := time.Now().Add(5 * time.Minute)
	atClaims := &сlaims{
		UserID: guid,
		UserIP: clientIP,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: atExpiration.Unix(),
			Id:        string(linkBytes),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, atClaims)
	atSigned, err := accessToken.SignedString(jwtKey)
	if err != nil {
		log.Printf("access token error: %v\n", err)
		return "", "", time.Now(), time.Now(), err
	}

	rTokens, err := pauthService.repo.GetUserRTokensByGUID(guid)
	if err != nil {
		log.Printf("failed to get refresh tokens from db: %v\n", err)
	}
	err = nil
	// save hashed refresh in db
	rtDataBaseString := CreateRefreshTokenFromData(rtClaims.UserID, rtClaims.Id, rtClaims.StandardClaims.ExpiresAt)
	rtHashed, err := bcrypt.GenerateFromPassword(rtDataBaseString, bcrypt.DefaultCost)

	rtHashedS := string(rtHashed)
	if err != nil {
		log.Printf("refresh token bcrypt error: %v\n", err)
		return "", "", time.Now(), time.Now(), err
	}
	rTokens = append(rTokens, rtHashedS)
	err = pauthService.repo.UpdateUserRefreshTokens(guid, rTokens)

	if err != nil {
		log.Printf("access token error: %v\n", err)
		return "", "", time.Now(), time.Now(), err
	}

	return atSigned, rtSigned, atExpiration, rtExpiration, nil
}

// usr token is jwt, db token is "custom" and bcrypted
// access is jwt, only on user
func (pauthService *AuthService) Refresh(usrRToken string, aToken string, clientIP string) (string, string, time.Time, time.Time, error) {

	rtClaims, err := getClaims(usrRToken)

	if err != nil {
		log.Printf("user's refresh token parsing failed: %s", err.Error())
		return "", "", time.Now(), time.Now(), err
	}

	atClaims, err := getClaims(aToken)

	if err != nil && err.Error() != "Token is not valid" {
		log.Printf("access token parsing failed: %s", err.Error())
		return "", "", time.Now(), time.Now(), err
	}
	// if anything is different in tokens, they are invalid
	if rtClaims.UserID != atClaims.UserID || rtClaims.Id != atClaims.Id {
		log.Printf("Unauthorized, invalid tokens")
		return "", "", time.Now(), time.Now(), err
	}

	dbRefreshTokens, err := pauthService.repo.GetUserRTokensByGUID(rtClaims.UserID)
	if err != nil || len(dbRefreshTokens) == 0 {
		log.Printf("Unauthorized %s", err.Error())
		return "", "", time.Now(), time.Now(), err
	}

	//checking db for refresh key
	Exists := false
	for i := len(dbRefreshTokens) - 1; i >= 0; i-- {
		err = bcrypt.CompareHashAndPassword([]byte(dbRefreshTokens[i]),
			CreateRefreshTokenFromData(rtClaims.UserID, rtClaims.Id, rtClaims.StandardClaims.ExpiresAt))

		if err == nil {
			Exists = true

			break
		}
	}
	if !Exists {
		log.Printf("refresh token not found in db: %s", err.Error())
		return "", "", time.Now(), time.Now(), errors.New("refresh not found in db")
	}

	// if IP is different, notify
	if clientIP != rtClaims.UserIP {
		from := "Notifier@example.com"

		user := "9c1d45eaf7af5b"
		password := "ad62926fa75d0f"

		uEmail, err := pauthService.repo.GetUsersEmailByGUID(rtClaims.UserID)
		if err != nil {
			log.Printf("failed to get users email from db: %s", err.Error())
		}

		to := []string{
			uEmail,
		}

		addr := "smtp.mailtrap.io:2525"
		host := "smtp.mailtrap.io"

		msg := []byte("From: Notifier@example.com\r\n" +
			"To: " + uEmail +
			"Subject: Test mail\r\n\r\n" +
			"Email body\r\n")

		auth := smtp.PlainAuth("", user, password, host)

		err = smtp.SendMail(addr, auth, from, to, msg)
		if err != nil {
			log.Printf("failed to notify user of IP change: %s", err.Error())
		}
	}

	//new access token
	atExpiration := time.Now().Add(5 * time.Minute)

	atClaims = &сlaims{
		UserID: atClaims.UserID,
		UserIP: clientIP,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: atExpiration.Unix(),
			Id:        rtClaims.Id,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, atClaims)
	atSigned, err := accessToken.SignedString(jwtKey)

	if err != nil {
		log.Printf("access token creation failed: %s", err.Error())
		return "", "", time.Now(), time.Now(), err
	}

	return atSigned, usrRToken, atExpiration, time.Unix(rtClaims.StandardClaims.ExpiresAt, 0), nil
}
