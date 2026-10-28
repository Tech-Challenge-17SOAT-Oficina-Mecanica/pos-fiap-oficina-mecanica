BEGIN;

ALTER TABLE movimentacao_estoque
    ADD COLUMN fornecedor_id UUID REFERENCES fornecedor (id);

CREATE TABLE auditoria_estoque (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_estoque_id UUID NOT NULL REFERENCES item_estoque (id),
    fornecedor_id UUID REFERENCES fornecedor (id),
    pedido_compra_id UUID REFERENCES pedido_compra (id),
    usuario_id UUID REFERENCES usuario (id),
    tipo_evento VARCHAR(100) NOT NULL,
    documento_origem VARCHAR(100) NOT NULL,
    dados JSONB NOT NULL DEFAULT '{}'::JSONB,
    ocorrido_em TIMESTAMPTZ NOT NULL,
    registrado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX ix_auditoria_estoque_item_ocorrido
    ON auditoria_estoque (item_estoque_id, ocorrido_em);

COMMIT;