BEGIN;

CREATE SEQUENCE IF NOT EXISTS servico_codigo_seq;

SELECT setval(
    'servico_codigo_seq',
    COALESCE(MAX(SUBSTRING(codigo FROM 5)::BIGINT), 1),
    COUNT(*) > 0
)
FROM servico;

ALTER TABLE servico
    ADD COLUMN IF NOT EXISTS usuario_criacao UUID REFERENCES usuario (id);

COMMIT;
