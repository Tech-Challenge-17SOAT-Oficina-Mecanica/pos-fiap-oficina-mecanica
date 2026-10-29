package seguranca

import "context"

type AutenticarUseCase interface {
	Execute(ctx context.Context, email, senha string) (string, error)
}
