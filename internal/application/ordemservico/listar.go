package ordemservico

import (
	"context"
	"errors"

	domainCliente "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	domainVeiculo "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

var (
	ErrStatusInvalido    = errors.New("status invalido")
	ErrDocumentoInvalido = errors.New("documento invalido")
	ErrPlacaInvalida     = errors.New("placa invalida")
)

type ListarInput struct {
	Status       string
	Documento    string
	Placa        string
	Limite       int
	Deslocamento int
}

type ResultadoListagem struct {
	Itens          []domain.ItemListagem
	TotalElementos int
}

type ListarRepository interface {
	Listar(ctx context.Context, filtros domain.FiltrosListagem, limite, deslocamento int) ([]domain.ItemListagem, int, error)
}

type Listar struct{ repository ListarRepository }

func NewListar(repository ListarRepository) Listar { return Listar{repository: repository} }

func (useCase Listar) Execute(ctx context.Context, input ListarInput) (ResultadoListagem, error) {
	if input.Status != "" && !domain.StatusListagemValido(input.Status) {
		return ResultadoListagem{}, ErrStatusInvalido
	}
	documento := input.Documento
	if documento != "" {
		normalizado, err := domainCliente.DocumentoParaConsulta(documento)
		if err != nil {
			return ResultadoListagem{}, ErrDocumentoInvalido
		}
		documento = normalizado
	}
	placa := input.Placa
	if placa != "" {
		normalizada, err := domainVeiculo.NormalizarPlaca(placa)
		if err != nil {
			return ResultadoListagem{}, ErrPlacaInvalida
		}
		placa = normalizada
	}

	itens, total, err := useCase.repository.Listar(ctx, domain.FiltrosListagem{Status: input.Status, Documento: documento, Placa: placa}, input.Limite, input.Deslocamento)
	if err != nil {
		return ResultadoListagem{}, err
	}
	return ResultadoListagem{Itens: itens, TotalElementos: total}, nil
}
