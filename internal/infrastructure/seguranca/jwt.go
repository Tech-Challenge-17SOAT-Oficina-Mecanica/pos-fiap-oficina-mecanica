package seguranca

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
)

var ErrTokenInvalido = errors.New("token inválido")

type jwtClaims struct {
	Escopos        []string `json:"escopos"`
	ClienteID      string   `json:"clienteId,omitempty"`
	OrdemServicoID string   `json:"ordemServicoId,omitempty"`
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

func (service JWT) GerarCliente(clienteID, ordemServicoID string) (string, error) {
	now := service.now()
	claims := jwtClaims{ClienteID: clienteID, OrdemServicoID: ordemServicoID, Escopos: []string{"orcamentos:ler"}, RegisteredClaims: jwt.RegisteredClaims{Subject: clienteID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour))}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(service.secret)
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
	return seguranca.Claims{UsuarioID: claims.Subject, ClienteID: claims.ClienteID, OrdemServicoID: claims.OrdemServicoID, Escopos: claims.Escopos}, nil
}
