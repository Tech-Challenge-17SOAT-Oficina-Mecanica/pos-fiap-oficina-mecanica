BEGIN;

ALTER TABLE ordem_servico
    ADD COLUMN problema_relatado_descricao TEXT,
    ADD COLUMN problema_relatado_observacoes TEXT,
    ADD COLUMN data_inicio_diagnostico TIMESTAMPTZ;

ALTER TABLE ordem_servico
    ADD CONSTRAINT ck_os_problema_relatado_completo CHECK (
        (problema_relatado_descricao IS NULL AND data_inicio_diagnostico IS NULL)
        OR
        (NULLIF(BTRIM(problema_relatado_descricao), '') IS NOT NULL AND data_inicio_diagnostico IS NOT NULL)
    );

COMMIT;
