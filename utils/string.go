package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lixiang4u/imago/models"
	"log"
	"time"
)

func StringMd5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func PasswordHash(password string) string {
	log.Println("[debug.password]", StringMd5(fmt.Sprintf("%s,%s", models.SECRET_KEY, password)))
	return StringMd5(fmt.Sprintf("%s,%s", models.SECRET_KEY, password))
}

func NewJwtAccessToken(id uint64, username string) (string, error) {
	claims := jwt.MapClaims{
		"id":   id,
		"name": username,
		"exp":  time.Now().Add(time.Hour * 1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(models.SECRET_KEY))
}

func NewJwtRefreshToken(id uint64, username string) (string, error) {
	claims := jwt.MapClaims{
		"id":   id,
		"name": username,
		"exp":  time.Now().Add(time.Hour * 24 * 30).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(models.SECRET_KEY))
}
