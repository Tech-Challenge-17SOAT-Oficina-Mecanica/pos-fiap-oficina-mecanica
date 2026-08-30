BEGIN;

ALTER TABLE ordem_servico
    ADD COLUMN data_entrada_fila TIMESTAMPTZ,
    ADD COLUMN version INTEGER NOT NULL DEFAULT 1,
    ADD CONSTRAINT ck_ordem_servico_version CHECK (version > 0);

COMMIT;
