package seguranca

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrTokenInvalido = errors.New("token inválido")

type Claims struct {
	Escopos []string `json:"escopos"`
	jwt.RegisteredClaims
}

type JWT struct {
	secret []byte
	now    func() time.Time
}

func NewJWT(secret string) (JWT, error) {
	if secret == "" {
		return JWT{}, errors.New("JWT_SECRET é obrigatório")
	}
	return JWT{secret: []byte(secret), now: time.Now}, nil
}

func (service JWT) Gerar(usuarioID string, escopos []string) (string, error) {
	now := service.now()
	claims := Claims{Escopos: escopos, RegisteredClaims: jwt.RegisteredClaims{Subject: usuarioID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(service.secret)
}

func (service JWT) Validar(raw string) (Claims, error) {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrTokenInvalido
		}
		return service.secret, nil
	})
	if err != nil || !token.Valid || claims.Subject == "" {
		return Claims{}, ErrTokenInvalido
	}
	return claims, nil
}
