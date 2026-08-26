BEGIN;

ALTER TABLE servico
    ADD COLUMN IF NOT EXISTS data_atualizacao TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS usuario_atualizacao UUID REFERENCES usuario (id);

COMMIT;
