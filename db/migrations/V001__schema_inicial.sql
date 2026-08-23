BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE categoria (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome VARCHAR(100) NOT NULL UNIQUE,
    ativa BOOLEAN NOT NULL DEFAULT TRUE,
    criada_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cliente (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome VARCHAR(200) NOT NULL,
    documento VARCHAR(14) NOT NULL,
    tipo_documento VARCHAR(4) NOT NULL,
    telefone VARCHAR(20),
    email VARCHAR(254),
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    inativado_em TIMESTAMPTZ,
    inativado_por UUID,
    version INTEGER NOT NULL DEFAULT 1,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_cliente_contato CHECK (telefone IS NOT NULL OR email IS NOT NULL),
    CONSTRAINT ck_cliente_version CHECK (version > 0)
);

CREATE UNIQUE INDEX ux_cliente_documento_ativo ON cliente (documento) WHERE ativo;

CREATE TABLE veiculo (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID NOT NULL REFERENCES cliente (id),
    placa VARCHAR(7) NOT NULL,
    marca VARCHAR(100) NOT NULL,
    modelo VARCHAR(100) NOT NULL,
    ano INTEGER NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    inativado_em TIMESTAMPTZ,
    inativado_por UUID,
    motivo_inativacao VARCHAR(200),
    version INTEGER NOT NULL DEFAULT 1,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_veiculo_ano CHECK (ano >= 1900),
    CONSTRAINT ck_veiculo_version CHECK (version > 0)
);

CREATE UNIQUE INDEX ux_veiculo_placa_ativa ON veiculo (placa) WHERE ativo;
CREATE INDEX ix_veiculo_cliente_id ON veiculo (cliente_id);

CREATE TABLE servico (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo VARCHAR(10) NOT NULL UNIQUE,
    nome VARCHAR(150) NOT NULL,
    nome_normalizado VARCHAR(150) NOT NULL,
    descricao TEXT,
    valor NUMERIC(12, 2) NOT NULL,
    tempo_estimado_minutos INTEGER NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    data_desativacao TIMESTAMPTZ,
    usuario_desativacao UUID,
    data_criacao TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT ck_servico_valor CHECK (valor >= 0),
    CONSTRAINT ck_servico_tempo CHECK (tempo_estimado_minutos >= 1),
    CONSTRAINT ck_servico_version CHECK (version > 0)
);

CREATE UNIQUE INDEX ux_servico_nome_normalizado_ativo ON servico (nome_normalizado) WHERE ativo;

CREATE TABLE item_estoque (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    categoria_id UUID NOT NULL REFERENCES categoria (id),
    tipo VARCHAR(7) NOT NULL,
    codigo VARCHAR(10) NOT NULL UNIQUE,
    nome VARCHAR(150) NOT NULL,
    descricao TEXT NOT NULL,
    descricao_normalizada TEXT NOT NULL,
    fabricante VARCHAR(150),
    unidade_medida VARCHAR(3) NOT NULL,
    saldo_fisico NUMERIC(14, 3) NOT NULL DEFAULT 0,
    saldo_reservado NUMERIC(14, 3) NOT NULL DEFAULT 0,
    estoque_minimo NUMERIC(14, 3) NOT NULL DEFAULT 0,
    preco_venda NUMERIC(12, 2),
    custo_unitario NUMERIC(12, 2),
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    data_desativacao TIMESTAMPTZ,
    usuario_desativacao UUID,
    version INTEGER NOT NULL DEFAULT 1,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_item_saldos CHECK (saldo_fisico >= 0 AND saldo_reservado >= 0 AND estoque_minimo >= 0),
    CONSTRAINT ck_item_version CHECK (version > 0)
);

CREATE UNIQUE INDEX ux_peca_categoria_descricao_ativa
    ON item_estoque (categoria_id, descricao_normalizada)
    WHERE ativo AND tipo = 'PECA';
CREATE UNIQUE INDEX ux_insumo_categoria_unidade_descricao_ativa
    ON item_estoque (categoria_id, unidade_medida, descricao_normalizada)
    WHERE ativo AND tipo = 'INSUMO';
CREATE INDEX ix_item_estoque_categoria_id ON item_estoque (categoria_id);

CREATE TABLE ordem_servico (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID NOT NULL REFERENCES cliente (id),
    veiculo_id UUID NOT NULL REFERENCES veiculo (id),
    placa_veiculo VARCHAR(7) NOT NULL,
    status VARCHAR(30) NOT NULL,
    custo_total_materiais NUMERIC(12, 2) NOT NULL DEFAULT 0,
    valor_final NUMERIC(12, 2),
    criada_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    iniciada_em TIMESTAMPTZ,
    finalizada_em TIMESTAMPTZ,
    observacoes_finalizacao TEXT,
    entregue_em TIMESTAMPTZ,
    cliente_retirada_id UUID REFERENCES cliente (id),
    responsavel_entrega_id UUID,
    observacoes_entrega TEXT,
    CONSTRAINT ck_os_custo CHECK (custo_total_materiais >= 0),
    CONSTRAINT ck_os_valor_final CHECK (valor_final IS NULL OR valor_final >= 0)
);

CREATE INDEX ix_ordem_servico_cliente_id ON ordem_servico (cliente_id);
CREATE INDEX ix_ordem_servico_veiculo_id ON ordem_servico (veiculo_id);
CREATE INDEX ix_ordem_servico_status ON ordem_servico (status);

CREATE TABLE problema_ordem_servico (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ordem_servico_id UUID NOT NULL REFERENCES ordem_servico (id),
    descricao TEXT NOT NULL,
    registrado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ix_problema_os_ordem_servico_id ON problema_ordem_servico (ordem_servico_id);

CREATE TABLE ordem_servico_servico (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ordem_servico_id UUID NOT NULL REFERENCES ordem_servico (id),
    servico_id UUID NOT NULL REFERENCES servico (id),
    descricao TEXT NOT NULL,
    valor_unitario NUMERIC(12, 2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    CONSTRAINT ck_os_servico_valor CHECK (valor_unitario >= 0)
);

CREATE INDEX ix_os_servico_ordem_servico_id ON ordem_servico_servico (ordem_servico_id);
CREATE INDEX ix_os_servico_servico_id ON ordem_servico_servico (servico_id);

CREATE TABLE ordem_servico_item (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ordem_servico_id UUID NOT NULL REFERENCES ordem_servico (id),
    item_estoque_id UUID NOT NULL REFERENCES item_estoque (id),
    quantidade_necessaria NUMERIC(14, 3) NOT NULL,
    quantidade_reservada NUMERIC(14, 3) NOT NULL DEFAULT 0,
    quantidade_consumida NUMERIC(14, 3) NOT NULL DEFAULT 0,
    valor_unitario NUMERIC(12, 2) NOT NULL,
    CONSTRAINT ck_os_item_quantidades CHECK (
        quantidade_necessaria > 0 AND quantidade_reservada >= 0 AND quantidade_consumida >= 0
    ),
    CONSTRAINT ck_os_item_valor CHECK (valor_unitario >= 0),
    CONSTRAINT ux_os_item UNIQUE (ordem_servico_id, item_estoque_id)
);

CREATE INDEX ix_os_item_item_estoque_id ON ordem_servico_item (item_estoque_id);

CREATE TABLE orcamento (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ordem_servico_id UUID NOT NULL REFERENCES ordem_servico (id),
    orcamento_original_id UUID REFERENCES orcamento (id),
    tipo_orcamento VARCHAR(15) NOT NULL,
    status VARCHAR(10) NOT NULL,
    estimativa_entrega_dias INTEGER,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    aprovado_em TIMESTAMPTZ,
    recusado_em TIMESTAMPTZ,
    data_atualizacao TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_orcamento_estimativa CHECK (estimativa_entrega_dias IS NULL OR estimativa_entrega_dias >= 0)
);

CREATE UNIQUE INDEX ux_orcamento_principal_por_os
    ON orcamento (ordem_servico_id) WHERE tipo_orcamento = 'PRINCIPAL';
CREATE INDEX ix_orcamento_ordem_servico_id ON orcamento (ordem_servico_id);

CREATE TABLE orcamento_item (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    orcamento_id UUID NOT NULL REFERENCES orcamento (id),
    servico_id UUID REFERENCES servico (id),
    item_estoque_id UUID REFERENCES item_estoque (id),
    tipo_item VARCHAR(10) NOT NULL,
    descricao TEXT NOT NULL,
    quantidade NUMERIC(14, 3) NOT NULL,
    valor_unitario NUMERIC(12, 2) NOT NULL,
    valor_total NUMERIC(12, 2) NOT NULL,
    CONSTRAINT ck_orcamento_item_valores CHECK (quantidade > 0 AND valor_unitario >= 0 AND valor_total >= 0)
);

CREATE INDEX ix_orcamento_item_orcamento_id ON orcamento_item (orcamento_id);

CREATE TABLE fornecedor (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    razao_social VARCHAR(200) NOT NULL,
    nome_fantasia VARCHAR(200),
    documento VARCHAR(14) NOT NULL,
    tipo_documento VARCHAR(4) NOT NULL,
    telefone VARCHAR(20),
    email VARCHAR(254),
    prazo_entrega_dias INTEGER NOT NULL DEFAULT 7,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    inativado_em TIMESTAMPTZ,
    inativado_por UUID,
    version INTEGER NOT NULL DEFAULT 1,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    usuario_atualizacao UUID,
    CONSTRAINT ck_fornecedor_contato CHECK (telefone IS NOT NULL OR email IS NOT NULL),
    CONSTRAINT ck_fornecedor_prazo CHECK (prazo_entrega_dias >= 0),
    CONSTRAINT ck_fornecedor_version CHECK (version > 0)
);

CREATE UNIQUE INDEX ux_fornecedor_documento_ativo ON fornecedor (documento) WHERE ativo;

CREATE TABLE pedido_compra (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fornecedor_id UUID NOT NULL REFERENCES fornecedor (id),
    numero VARCHAR(30) NOT NULL UNIQUE,
    status VARCHAR(10) NOT NULL,
    solicitado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    recebido_em TIMESTAMPTZ
);

CREATE INDEX ix_pedido_compra_fornecedor_id ON pedido_compra (fornecedor_id);
CREATE INDEX ix_pedido_compra_status ON pedido_compra (status);

CREATE TABLE pedido_compra_item (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pedido_compra_id UUID NOT NULL REFERENCES pedido_compra (id),
    item_estoque_id UUID NOT NULL REFERENCES item_estoque (id),
    quantidade_necessaria NUMERIC(14, 3) NOT NULL,
    quantidade_pedida NUMERIC(14, 3) NOT NULL,
    quantidade_reservada NUMERIC(14, 3) NOT NULL DEFAULT 0,
    quantidade_recebida NUMERIC(14, 3) NOT NULL DEFAULT 0,
    custo_unitario NUMERIC(12, 2),
    CONSTRAINT ck_pedido_item_quantidades CHECK (
        quantidade_necessaria > 0 AND quantidade_pedida >= quantidade_necessaria
        AND quantidade_reservada >= 0 AND quantidade_recebida >= 0
    ),
    CONSTRAINT ck_pedido_item_custo CHECK (custo_unitario IS NULL OR custo_unitario >= 0),
    CONSTRAINT ux_pedido_item UNIQUE (pedido_compra_id, item_estoque_id)
);

CREATE INDEX ix_pedido_compra_item_item_estoque_id ON pedido_compra_item (item_estoque_id);

CREATE TABLE pedido_compra_item_os (
    pedido_compra_item_id UUID NOT NULL REFERENCES pedido_compra_item (id),
    ordem_servico_item_id UUID NOT NULL REFERENCES ordem_servico_item (id),
    quantidade_atendida NUMERIC(14, 3) NOT NULL,
    PRIMARY KEY (pedido_compra_item_id, ordem_servico_item_id),
    CONSTRAINT ck_pedido_item_os_quantidade CHECK (quantidade_atendida > 0)
);

CREATE TABLE reserva_estoque (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ordem_servico_item_id UUID NOT NULL REFERENCES ordem_servico_item (id),
    item_estoque_id UUID NOT NULL REFERENCES item_estoque (id),
    pedido_compra_item_id UUID REFERENCES pedido_compra_item (id),
    quantidade NUMERIC(14, 3) NOT NULL,
    status VARCHAR(10) NOT NULL,
    reservada_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    liberada_em TIMESTAMPTZ,
    CONSTRAINT ck_reserva_quantidade CHECK (quantidade > 0)
);

CREATE INDEX ix_reserva_os_item_id ON reserva_estoque (ordem_servico_item_id);
CREATE INDEX ix_reserva_item_estoque_id ON reserva_estoque (item_estoque_id);

CREATE TABLE movimentacao_estoque (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_estoque_id UUID NOT NULL REFERENCES item_estoque (id),
    ordem_servico_id UUID REFERENCES ordem_servico (id),
    reserva_estoque_id UUID REFERENCES reserva_estoque (id),
    pedido_compra_id UUID REFERENCES pedido_compra (id),
    tipo VARCHAR(20) NOT NULL,
    quantidade NUMERIC(14, 3) NOT NULL,
    custo_unitario NUMERIC(12, 2),
    documento_origem VARCHAR(100),
    ocorrida_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_movimentacao_quantidade CHECK (quantidade > 0),
    CONSTRAINT ck_movimentacao_custo CHECK (custo_unitario IS NULL OR custo_unitario >= 0)
);

CREATE UNIQUE INDEX ux_movimentacao_documento_item
    ON movimentacao_estoque (documento_origem, item_estoque_id)
    WHERE documento_origem IS NOT NULL;
CREATE INDEX ix_movimentacao_item_estoque_id ON movimentacao_estoque (item_estoque_id);
CREATE INDEX ix_movimentacao_ordem_servico_id ON movimentacao_estoque (ordem_servico_id);

CREATE TABLE auditoria_ordem_servico (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ordem_servico_id UUID NOT NULL REFERENCES ordem_servico (id),
    usuario_id UUID,
    agregado VARCHAR(100) NOT NULL,
    agregado_id UUID NOT NULL,
    tipo_evento VARCHAR(100) NOT NULL,
    dados JSONB NOT NULL DEFAULT '{}'::JSONB,
    metadados JSONB NOT NULL DEFAULT '{}'::JSONB,
    ocorrido_em TIMESTAMPTZ NOT NULL,
    registrado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ix_auditoria_os_ordem_servico_ocorrido ON auditoria_ordem_servico (ordem_servico_id, ocorrido_em);

CREATE TABLE chave_idempotencia (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chave VARCHAR(255) NOT NULL,
    operacao VARCHAR(100) NOT NULL,
    hash_requisicao VARCHAR(128) NOT NULL,
    status_resposta INTEGER NOT NULL,
    resposta JSONB,
    processada_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ux_chave_idempotencia_operacao_chave UNIQUE (operacao, chave)
);

COMMIT;
