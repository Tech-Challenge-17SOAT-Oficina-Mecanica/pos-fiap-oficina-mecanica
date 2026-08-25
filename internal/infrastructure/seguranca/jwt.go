package seguranca

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
)

var ErrTokenInvalido = errors.New("token inválido")

type jwtClaims struct {
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
	claims := jwtClaims{Escopos: escopos, RegisteredClaims: jwt.RegisteredClaims{Subject: usuarioID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(service.secret)
}

func (service JWT) Autenticar(raw string) (string, []string, error) {
	claims, err := service.Validar(raw)
	if err != nil {
		return "", nil, err
	}
	return claims.UsuarioID, claims.Escopos, nil
}

func (service JWT) Validar(raw string) (seguranca.Claims, error) {
	claims := jwtClaims{}
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrTokenInvalido
		}
		return service.secret, nil
	})
	if err != nil || !token.Valid || claims.Subject == "" {
		return seguranca.Claims{}, ErrTokenInvalido
	}
	return seguranca.Claims{UsuarioID: claims.Subject, Escopos: claims.Escopos}, nil
}
