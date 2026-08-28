BEGIN;

-- Auditoria da atualizacao, no mesmo formato ja usado por fornecedor e servico.
ALTER TABLE item_estoque
    ADD COLUMN IF NOT EXISTS data_atualizacao TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS usuario_atualizacao UUID REFERENCES usuario (id);

-- Historico de preco do item. A D-14 assume que ele existe para decidir o custo do
-- insumo pela ultima entrada, e a atualizacao de peca grava aqui quando o preco muda.
CREATE TABLE IF NOT EXISTS historico_preco_item (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_estoque_id UUID NOT NULL REFERENCES item_estoque (id),
    preco_anterior NUMERIC(12, 2),
    preco_novo NUMERIC(12, 2) NOT NULL,
    usuario_id UUID REFERENCES usuario (id),
    registrado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_historico_preco_novo CHECK (preco_novo >= 0),
    CONSTRAINT ck_historico_preco_anterior CHECK (preco_anterior IS NULL OR preco_anterior >= 0)
);

CREATE INDEX IF NOT EXISTS ix_historico_preco_item_id
    ON historico_preco_item (item_estoque_id, registrado_em DESC);

COMMIT;
