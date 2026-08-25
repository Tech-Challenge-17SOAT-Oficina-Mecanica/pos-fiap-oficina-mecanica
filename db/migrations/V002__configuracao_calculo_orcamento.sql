BEGIN;

CREATE TABLE IF NOT EXISTS configuracao_oficina (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    capacidade_diaria_os INTEGER NOT NULL,
    minutos_produtivos_dia INTEGER NOT NULL,
    atualizado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ck_configuracao_oficina_unica CHECK (id = 1),
    CONSTRAINT ck_configuracao_capacidade CHECK (capacidade_diaria_os > 0),
    CONSTRAINT ck_configuracao_minutos CHECK (minutos_produtivos_dia > 0)
);

INSERT INTO configuracao_oficina (id, capacidade_diaria_os, minutos_produtivos_dia)
VALUES (1, 3, 480)
ON CONFLICT (id) DO NOTHING;

COMMIT;
