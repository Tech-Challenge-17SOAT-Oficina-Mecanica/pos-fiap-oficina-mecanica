package peca

import (
	"context"
	"errors"
	"testing"
)

const idValido = "10000000-0000-0000-0000-000000000001"

type processamentoFake struct {
	solicitacao SolicitacaoCompraReserva
	resultado   ResultadoCompraReserva
	err         error
}

func (fake *processamentoFake) SolicitarCompraEReservar(_ context.Context, solicitacao SolicitacaoCompraReserva) (ResultadoCompraReserva, error) {
	fake.solicitacao = solicitacao
	return fake.resultado, fake.err
}

func TestSolicitarCompraEReservarPecasValidaEDelega(t *testing.T) {
	fake := &processamentoFake{resultado: ResultadoCompraReserva{OrdemServicoID: idValido}}
	resultado, err := NewSolicitarCompraEReservarPecas(fake).Execute(context.Background(), solicitacaoValida())

	if err != nil || resultado.OrdemServicoID != idValido {
		t.Fatalf("resultado=%+v erro=%v", resultado, err)
	}
	if fake.solicitacao.OrdemServicoID != idValido || fake.solicitacao.Itens[0].Quantidade != 2 {
		t.Fatalf("solicitacao nao delegada: %+v", fake.solicitacao)
	}
}

func TestSolicitarCompraEReservarPecasValidacoes(t *testing.T) {
	casos := []struct {
		nome  string
		mudar func(*SolicitacaoCompraReserva)
		erro  error
	}{
		{"chave ausente", func(s *SolicitacaoCompraReserva) { s.IdempotencyKey = "" }, ErrIdempotencyKeyObrigatoria},
		{"os invalida", func(s *SolicitacaoCompraReserva) { s.OrdemServicoID = "x" }, ErrIdentificadorInvalido},
		{"fornecedor invalido", func(s *SolicitacaoCompraReserva) { s.FornecedorID = "x" }, ErrFornecedorIdentificador},
		{"sem item", func(s *SolicitacaoCompraReserva) { s.Itens = nil }, ErrItemObrigatorio},
		{"item invalido", func(s *SolicitacaoCompraReserva) { s.Itens[0].ItemID = "x" }, ErrItemIdentificador},
		{"quantidade invalida", func(s *SolicitacaoCompraReserva) { s.Itens[0].Quantidade = 0 }, ErrQuantidadeProcessamento},
		{"item repetido", func(s *SolicitacaoCompraReserva) { s.Itens = append(s.Itens, s.Itens[0]) }, ErrItemRepetido},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			solicitacao := solicitacaoValida()
			caso.mudar(&solicitacao)
			_, err := NewSolicitarCompraEReservarPecas(&processamentoFake{}).Execute(context.Background(), solicitacao)
			if !errors.Is(err, caso.erro) {
				t.Fatalf("erro=%v, esperado %v", err, caso.erro)
			}
		})
	}
}

func TestSolicitarCompraEReservarPecasPropagaErro(t *testing.T) {
	esperado := errors.New("falha")
	_, err := NewSolicitarCompraEReservarPecas(&processamentoFake{err: esperado}).Execute(context.Background(), solicitacaoValida())
	if !errors.Is(err, esperado) {
		t.Fatalf("erro=%v", err)
	}
}

func solicitacaoValida() SolicitacaoCompraReserva {
	return SolicitacaoCompraReserva{
		IdempotencyKey: "10000000-0000-0000-0000-000000000001",
		HashRequisicao: "hash",
		OrdemServicoID: idValido,
		FornecedorID:   "20000000-0000-0000-0000-000000000001",
		Itens:          []ItemProcessamento{{ItemID: "30000000-0000-0000-0000-000000000001", Quantidade: 2}},
	}
}
