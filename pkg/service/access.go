package service

import (
	"crypto/rand"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

// returns a pair of access and refresh tokens

// usr token is jwt, db token is "custom" and bcrypted
// access is jwt, only on user
func (authService *AuthService) Access(guid string, clientIP string) (string, string, time.Time, time.Time, error) {

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
	//jwt is base64 and cant be changed by user

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
	// access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, atClaims)
	atSigned, err := accessToken.SignedString(jwtKey)
	if err != nil {
		log.Printf("access token error: %v\n", err)
		return "", "", time.Now(), time.Now(), err
	}

	rTokens, err := authService.repo.GetUserRTokensByGUID(guid)
	if err != nil {
		log.Printf("failed to get refresh tokens from db: %v\n", err)
	}
	// Create a server side refresh token, store as bcrypt hash in DB
	rtDataBaseString := CreateRefreshTokenFromData(rtClaims.UserID, rtClaims.Id, rtClaims.StandardClaims.ExpiresAt)
	rtHashed, err := bcrypt.GenerateFromPassword(rtDataBaseString, bcrypt.DefaultCost)

	rtHashedS := string(rtHashed)
	if err != nil {
		log.Printf("refresh token bcrypt error: %v\n", err)
		return "", "", time.Now(), time.Now(), err
	}
	rTokens = append(rTokens, rtHashedS)
	err = authService.repo.UpdateUserRefreshTokens(guid, rTokens)

	if err != nil {
		log.Printf("access token error: %v\n", err)
		return "", "", time.Now(), time.Now(), err
	}

	return atSigned, rtSigned, atExpiration, rtExpiration, nil
}
