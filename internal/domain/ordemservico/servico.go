package ordemservico

import (
	"errors"
	"strings"
)

var ErrServicosObrigatorios = errors.New("informe ao menos um servico")
var ErrServicoDuplicado = errors.New("servico duplicado no orcamento")
var ErrStatusNaoPermiteServico = errors.New("status da ordem de servico nao permite registrar servicos")
var ErrOrcamentoAplicavelNaoEncontrado = errors.New("orcamento aplicavel nao encontrado para a ordem de servico")
var ErrServicoNaoEncontrado = errors.New("servico nao encontrado")
var ErrServicoInativo = errors.New("servico inativo")

type ServicoCadastro struct {
	ServicoID  string
	Observacao string
}

type ServicoRegistrado struct {
	ServicoID     string
	Descricao     string
	ValorUnitario float64
	Observacao    string
}

func NovosServicosCadastro(servicos []ServicoCadastro) ([]ServicoCadastro, error) {
	if len(servicos) == 0 {
		return nil, ErrServicosObrigatorios
	}
	ids := make(map[string]struct{}, len(servicos))
	result := make([]ServicoCadastro, 0, len(servicos))
	for _, servico := range servicos {
		servico.ServicoID = strings.TrimSpace(servico.ServicoID)
		servico.Observacao = strings.TrimSpace(servico.Observacao)
		if _, found := ids[servico.ServicoID]; found {
			return nil, ErrServicoDuplicado
		}
		ids[servico.ServicoID] = struct{}{}
		result = append(result, servico)
	}
	return result, nil
}

func TipoOrcamentoParaServico(status string) (string, error) {
	switch status {
	case StatusEmDiagnostico:
		return OrcamentoPrincipal, nil
	case StatusEmExecucao:
		return OrcamentoComplementar, nil
	default:
		return "", ErrStatusNaoPermiteServico
	}
}
