package repository

import (
	"fmt"
	"log"

	goproj "github.com/gfifgfifofich/GoProj"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type AuthDb struct {
	DB *sqlx.DB
}

func NewAuthDb(DB *sqlx.DB) *AuthDb {
	return &AuthDb{DB: DB}
}
func (AuthDB *AuthDb) CreateUser(user goproj.User) (string, error) {
	query := fmt.Sprintf("INSERT INTO %s (name,guid,password_hash) values($1,$2,$3) RETURNING id", usersTable)
	var id int
	guid, err := uuid.NewV4()
	if err != nil {
		log.Print("Failed to generate guid")
	}
	row := AuthDB.DB.QueryRow(query, user.Email, guid, user.Password)

	if err := row.Scan(&id); err != nil {
		return "", err
	}
	return guid.String(), nil
}

// converts a string array into single string, that can be sent to Postgres and stored as array there
func ConvertToPostgresStringArray(strarr []string) (string, error) {
	//format  '{"text1","text2"}';
	first := true
	outstr := "{"
	for i := 0; i < len(strarr); i++ {
		if !first {
			outstr += ","
		}
		first = false
		outstr += "\""
		if len(strarr[i]) != 0 {
			outstr += strarr[i]
		} else {
			i++
			if i < len(strarr) {
				outstr += strarr[i]
			}
		}
		outstr += "\""
	}
	outstr += "}"
	return outstr, nil
}

// converts Postgress array (single string) back into string array
func ConvertFromPostgresStringArray(str string) ([]string, error) {
	//format  {text1,text2};
	strarr := []string{}
	var tmp string
	tmp = ""
	for i := 0; i < len(str); i++ {
		if str[i] == ',' || str[i] == '}' {
			if len(tmp) != 0 {
				strarr = append(strarr, tmp)
			}
			tmp = ""
		} else if str[i] != '{' {
			tmp += string(str[i])
		}
	}
	return strarr, nil
}

func (pAuthDB *AuthDb) GetUserRTokensByGUID(guid string) ([]string, error) {
	query := fmt.Sprintf("SELECT refreshtokens FROM %s where guid = $1", usersTable)
	row := pAuthDB.DB.QueryRow(query, guid)
	var str string
	if err := row.Scan(&str); err != nil {
		return []string{}, err
	}
	rTokens, err := ConvertFromPostgresStringArray(str)
	if err != nil {
		log.Print("Failed to read tokens from db")
		return rTokens, err
	}
	if len(rTokens) == 1 && len(rTokens[0]) == 2 {
		return []string{}, err
	}
	return rTokens, nil
}

func (pAuthDB *AuthDb) GetUsersEmailByGUID(guid string) (string, error) {
	query := fmt.Sprintf("SELECT name FROM %s where guid = $1", usersTable)
	row := pAuthDB.DB.QueryRow(query, guid)
	var str string
	if err := row.Scan(&str); err != nil {
		log.Print("Failed to read Email from db")
		return "", err
	}
	return str, nil
}

func (pAuthDB *AuthDb) UpdateUserRefreshTokens(guid string, rTokens []string) error {
	query := fmt.Sprintf("UPDATE %s SET refreshtokens = $1 where guid = $2", usersTable)
	str, err := ConvertToPostgresStringArray(rTokens)
	if err != nil {
		log.Print("No token given")
	}
	pAuthDB.DB.QueryRow(query, str, guid).Scan(&str)

	return nil
}
