package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"
	"unicode/utf8"
	"villcore.com/admin/db"
	"villcore.com/common/model"
)

func init() {
	log.SetOutput(os.Stdout)
}

func DoLogin(username, password string) (string, error) {
	log.Printf("Do login with param, username %v, password %v, remever %v \n", username, password)
	users := make([]model.JobUser, 0)
	err := db.DbEngine.Table("xxl_job_user").Where("username = ?", username).Find(&users)
	if err != nil || len(users) <= 0 {
		return "", err
	}

	user := users[0]
	hash := md5.New()
	_, err = hash.Write([]byte(password))
	if err != nil {
		return "", err
	}
	passwordMd5 := hex.EncodeToString(hash.Sum(nil))

	if passwordMd5 != user.Password {
		return "", errors.New("login_param_unvalid")
	}

	// login success
	token, err := makeToken(&user)
	if err != nil {
		return "", err
	}
	return token, nil
}

func makeToken(user *model.JobUser) (string, error) {
	bytes, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GetUserFromToken(token string) (*model.JobUser, error) {
	if utf8.RuneCountInString(token) <= 0 {
		return nil, errors.New("Invalid param token ")
	}

	bytes, err := hex.DecodeString(token)
	if err != nil {
		return nil, err
	}

	user := new(model.JobUser)
	err = json.Unmarshal(bytes, user)
	if err != nil {
		log.Println("Parse token ", token, " error ", err)
		return nil, err
	}

	users := make([]model.JobUser, 0)
	err = db.DbEngine.Table("xxl_job_user").Where("username = ?", user.Username).Find(&users)
	if err != nil || len(users) <= 0 {
		return nil, err
	}
	return &users[0], nil

}
