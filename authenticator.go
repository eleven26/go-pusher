package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type Authenticator interface {
	// Authenticate 验证请求是否合法，第一个返回值为用户 id，第二个返回值为错误
	Authenticate(r *http.Request) (string, error)
}

var _ Authenticator = &JWTAuthenticator{}

type JWTAuthenticator struct{}

func (J *JWTAuthenticator) Authenticate(r *http.Request) (string, error) {
	j := newJwt(r.FormValue("token"))
	return j.parse()
}

// Jwt 需要配置 JWT_SECRET 环境变量
type Jwt struct {
	Secret string
	Token  string
}

func newJwt(token string) *Jwt {
	return &Jwt{
		Token:  token,
		Secret: os.Getenv("JWT_SECRET"),
	}
}

func (j *Jwt) parse() (string, error) {
	token, err := jwt.Parse(j.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.Secret), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if uid, ok := claims["uid"].(string); ok {
			return uid, nil
		}
		return "", fmt.Errorf("uid should be string")
	}

	return "", fmt.Errorf("jwt parse error")
}
