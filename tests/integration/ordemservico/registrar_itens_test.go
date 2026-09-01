package integration_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	domainOS "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/ordemservico"
)

func TestRegistrarPecasEInsumosNoOrcamento(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL nao configurada")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	suffix := time.Now().UnixNano() & 0xffffffffffff
	id := func(prefix string) string { return fmt.Sprintf(prefix+"%012x", suffix) }
	clienteID, veiculoID := id("a1000000-0000-0000-0000-"), id("a2000000-0000-0000-0000-")
	categoriaID := id("a3000000-0000-0000-0000-")
	osDiagnosticoID, osExecucaoID, osFechadaID := id("a4000000-0000-0000-0000-"), id("a5000000-0000-0000-0000-"), id("a6000000-0000-0000-0000-")
	orcamentoDiagnosticoID, orcamentoPrincipalID := id("a7000000-0000-0000-0000-"), id("a8000000-0000-0000-0000-")
	pecaID, insumoID, pecaInativaID, pecaSemValorID := id("a9000000-0000-0000-0000-"), id("aa000000-0000-0000-0000-"), id("ab000000-0000-0000-0000-"), id("ac000000-0000-0000-0000-")
	placa := placaMercosul("ITM", suffix)

	if _, err = db.Exec(ctx, "INSERT INTO cliente (id,nome,documento,tipo_documento,telefone) VALUES ($1,'Cliente Itens',$2,'CPF','11999999999')", clienteID, cpfValido(suffix%1000000000)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO veiculo (id,cliente_id,placa,marca,modelo,ano) VALUES ($1,$2,$3,'Teste','Teste',2024)", veiculoID, clienteID, placa); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO categoria (id,nome,ativa) VALUES ($1,$2,true)`, categoriaID, fmt.Sprintf("Categoria Itens %x", suffix)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO ordem_servico (id,cliente_id,veiculo_id,placa_veiculo,status) VALUES
		($1,$2,$3,$4,'EM_DIAGNOSTICO'),($5,$2,$3,$4,'EM_EXECUCAO'),($6,$2,$3,$4,'ENTREGUE')`,
		osDiagnosticoID, clienteID, veiculoID, placa, osExecucaoID, osFechadaID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, `INSERT INTO orcamento (id,ordem_servico_id,tipo_orcamento,status) VALUES
		($1,$2,'PRINCIPAL','CRIADO'),($3,$4,'PRINCIPAL','APROVADO')`,
		orcamentoDiagnosticoID, osDiagnosticoID, orcamentoPrincipalID, osExecucaoID); err != nil {
		t.Fatal(err)
	}
	codigoBase := suffix & 0xffffff
	codigoPeca := fmt.Sprintf("P%06x1", codigoBase)
	codigoInsumo := fmt.Sprintf("I%06x1", codigoBase)
	codigoPecaInativa := fmt.Sprintf("P%06x2", codigoBase)
	codigoPecaSemValor := fmt.Sprintf("P%06x3", codigoBase)
	if _, err = db.Exec(ctx, `INSERT INTO item_estoque
		(id,categoria_id,tipo,codigo,nome,descricao,descricao_normalizada,unidade_medida,saldo_fisico,saldo_reservado,preco_venda,custo_unitario,ativo)
		VALUES
		($1,$2,'PECA',$3,'Pastilha','Pastilha de freio',$4,'UN',10,0,120,80,true),
		($5,$2,'INSUMO',$6,'Oleo','Oleo 5w30',$7,'L',10,0,NULL,35,true),
		($8,$2,'PECA',$9,'Filtro inativo','Filtro inativo',$10,'UN',10,0,50,30,false),
		($11,$2,'PECA',$12,'Peca sem valor','Peca sem valor',$13,'UN',10,0,NULL,30,true)`,
		pecaID, categoriaID, codigoPeca, codigoPeca,
		insumoID, codigoInsumo, codigoInsumo,
		pecaInativaID, codigoPecaInativa, codigoPecaInativa,
		pecaSemValorID, codigoPecaSemValor, codigoPecaSemValor); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM auditoria_ordem_servico WHERE ordem_servico_id IN ($1,$2,$3)", osDiagnosticoID, osExecucaoID, osFechadaID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento_item WHERE orcamento_id IN (SELECT id FROM orcamento WHERE ordem_servico_id IN ($1,$2,$3))", osDiagnosticoID, osExecucaoID, osFechadaID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico_item WHERE ordem_servico_id IN ($1,$2,$3)", osDiagnosticoID, osExecucaoID, osFechadaID)
		_, _ = db.Exec(ctx, "DELETE FROM orcamento WHERE ordem_servico_id IN ($1,$2,$3)", osDiagnosticoID, osExecucaoID, osFechadaID)
		_, _ = db.Exec(ctx, "DELETE FROM ordem_servico WHERE id IN ($1,$2,$3)", osDiagnosticoID, osExecucaoID, osFechadaID)
		_, _ = db.Exec(ctx, "DELETE FROM item_estoque WHERE id IN ($1,$2,$3,$4)", pecaID, insumoID, pecaInativaID, pecaSemValorID)
		_, _ = db.Exec(ctx, "DELETE FROM categoria WHERE id=$1", categoriaID)
		_, _ = db.Exec(ctx, "DELETE FROM veiculo WHERE id=$1", veiculoID)
		_, _ = db.Exec(ctx, "DELETE FROM cliente WHERE id=$1", clienteID)
	})

	useCase := application.NewRegistrarItens(infrastructure.NewPostgresRepository(db))
	usuarioID := "90000000-0000-0000-0000-000000000001"

	resultado, err := useCase.Execute(ctx, application.RegistrarInput{
		OSID: osDiagnosticoID, Tipo: domainEstoque.TipoPeca, UsuarioID: usuarioID,
		Itens: []application.ItemInput{{ItemID: pecaID, Quantidade: 2}},
	})
	if err != nil || resultado.ValorOrcamento != 240 || len(resultado.ItensRegistrados) != 1 {
		t.Fatalf("resultado=%+v err=%v", resultado, err)
	}

	resultado, err = useCase.Execute(ctx, application.RegistrarInput{
		OSID: osExecucaoID, Tipo: domainEstoque.TipoInsumo, UsuarioID: usuarioID,
		Itens: []application.ItemInput{{ItemID: insumoID, Quantidade: 1.5}},
	})
	if err != nil || resultado.TipoOrcamento != "COMPLEMENTAR" || resultado.OrcamentoOriginal != orcamentoPrincipalID {
		t.Fatalf("complementar=%+v err=%v", resultado, err)
	}

	var quantidadeNecessaria float64
	if err = db.QueryRow(ctx, "SELECT quantidade_necessaria FROM ordem_servico_item WHERE ordem_servico_id=$1 AND item_estoque_id=$2", osDiagnosticoID, pecaID).Scan(&quantidadeNecessaria); err != nil || quantidadeNecessaria != 2 {
		t.Fatalf("quantidade=%.3f err=%v", quantidadeNecessaria, err)
	}

	for _, test := range []struct {
		name  string
		input application.RegistrarInput
		want  error
	}{
		{"os ausente", application.RegistrarInput{OSID: "a4000000-0000-0000-0000-000000000099", Tipo: domainEstoque.TipoPeca, Itens: []application.ItemInput{{ItemID: pecaID, Quantidade: 1}}}, application.ErrOSNaoEncontrada},
		{"status invalido", application.RegistrarInput{OSID: osFechadaID, Tipo: domainEstoque.TipoPeca, Itens: []application.ItemInput{{ItemID: pecaID, Quantidade: 1}}}, domainOS.ErrStatusNaoPermiteItens},
		{"item ausente", application.RegistrarInput{OSID: osDiagnosticoID, Tipo: domainEstoque.TipoPeca, Itens: []application.ItemInput{{ItemID: "a9000000-0000-0000-0000-000000000099", Quantidade: 1}}}, application.ErrItemNaoEncontrado},
		{"item inativo", application.RegistrarInput{OSID: osDiagnosticoID, Tipo: domainEstoque.TipoPeca, Itens: []application.ItemInput{{ItemID: pecaInativaID, Quantidade: 1}}}, application.ErrItemInativo},
		{"tipo divergente", application.RegistrarInput{OSID: osDiagnosticoID, Tipo: domainEstoque.TipoInsumo, Itens: []application.ItemInput{{ItemID: pecaID, Quantidade: 1}}}, domainEstoque.ErrTipoItemInvalido},
		{"sem valor", application.RegistrarInput{OSID: osDiagnosticoID, Tipo: domainEstoque.TipoPeca, Itens: []application.ItemInput{{ItemID: pecaSemValorID, Quantidade: 1}}}, application.ErrItemSemValor},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := useCase.Execute(ctx, test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("err=%v esperado=%v", err, test.want)
			}
		})
	}
}
