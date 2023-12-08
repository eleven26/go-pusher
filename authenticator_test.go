package main

import (
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const (
	_testToken  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjE0NDQ0Nzg0MDAsInVpZCI6IjEyMyJ9.9yJ-ABQGJkdnDqHo-wV-vojQFEQGt-I0dyva1w6EQ7E"
	_testSecret = "secret"
	_testUid    = "123"
)

func TestGenerateToken(t *testing.T) {
	// 注意：在生产使用中，还需要加上有效时间等限制
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": _testUid,
		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString([]byte(_testSecret))
	assert.Nil(t, err)
	assert.Equal(t, _testToken, tokenString)
}

func TestJWTAuthenticator(t *testing.T) {
	_ = os.Setenv("JWT_SECRET", "secret")

	var a Authenticator = &JWTAuthenticator{}
	u, _ := url.Parse("ws://localhost:8181/ws?token=" + _testToken)
	uid, err := a.Authenticate(&http.Request{URL: u})

	assert.Nil(t, err)
	assert.Equal(t, _testUid, uid)
}
