BEGIN;

ALTER TABLE ordem_servico
    ADD COLUMN IF NOT EXISTS data_entrada_fila TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1,
    ADD CONSTRAINT ck_ordem_servico_version CHECK (version > 0);

COMMIT;
