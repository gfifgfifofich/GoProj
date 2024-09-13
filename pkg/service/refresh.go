package service

import (
	"errors"
	"log"
	"net/smtp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func (authService *AuthService) Refresh(usrRefreshToken string, AccessToken string, clientIP string) (string, time.Time, error) {

	RefreshClaims, err := getClaims(usrRefreshToken)
	if err != nil {
		log.Printf("user's refresh token parsing failed: %s", err.Error())
		return "", time.Now(), err
	}

	AccessClaims, err := getClaims(AccessToken)
	if err != nil && err.Error() != "Token is not valid" {
		log.Printf("access token parsing failed: %s", err.Error())
		return "", time.Now(), err
	}

	// if anything is different in tokens, they are invalid
	if RefreshClaims.UserID != AccessClaims.UserID || RefreshClaims.Id != AccessClaims.Id {
		// dont forget to set err to something, or it will return nil and it will bypass all checks if tokens are invalid
		return "", time.Now(), errors.New("invalid tokens")
	}

	dbRefreshTokens, err := authService.repo.GetUserRTokensByGUID(RefreshClaims.UserID)
	if err != nil || len(dbRefreshTokens) == 0 {
		log.Printf("Unauthorized %s", err.Error())
		return "", time.Now(), err
	}

	rtoken := CreateRefreshTokenFromData(RefreshClaims.UserID, RefreshClaims.Id, RefreshClaims.StandardClaims.ExpiresAt)
	//checking db for refresh key
	Exists := false
	for i := len(dbRefreshTokens) - 1; i >= 0; i-- {
		err = bcrypt.CompareHashAndPassword([]byte(dbRefreshTokens[i]), rtoken)
		if err == nil {
			Exists = true
			break
		}
	}
	if !Exists {
		log.Printf("refresh token not found in db: %s", err.Error())
		return "", time.Now(), errors.New("refresh not found in db")
	}
	// if IP is different, notify
	if clientIP != RefreshClaims.UserIP {
		from := "Notifier@example.com"

		user := "9c1d45eaf7af5b"
		password := "ad62926fa75d0f"

		uEmail, err := authService.repo.GetUsersEmailByGUID(RefreshClaims.UserID)
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
	AccessTokenExpiration := time.Now().Add(5 * time.Minute)
	AccessClaims = &сlaims{
		UserID: AccessClaims.UserID,
		UserIP: clientIP,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: AccessTokenExpiration.Unix(),
			Id:        RefreshClaims.Id,
		},
	}
	NewAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, AccessClaims)
	NewAccessTokenSigned, err := NewAccessToken.SignedString(jwtKey)
	if err != nil {
		log.Printf("access token creation failed: %s", err.Error())
		return "", time.Now(), err
	}

	return NewAccessTokenSigned, AccessTokenExpiration, nil
}
