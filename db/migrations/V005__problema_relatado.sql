BEGIN;

ALTER TABLE problema_ordem_servico
    ADD COLUMN tipo VARCHAR(10) NOT NULL DEFAULT 'ENCONTRADO',
    ADD COLUMN observacoes TEXT;

ALTER TABLE problema_ordem_servico
    ADD CONSTRAINT ck_problema_os_tipo CHECK (tipo IN ('RELATADO', 'ENCONTRADO'));

CREATE UNIQUE INDEX ux_problema_relatado_por_os
    ON problema_ordem_servico (ordem_servico_id)
    WHERE tipo = 'RELATADO';

COMMIT;

