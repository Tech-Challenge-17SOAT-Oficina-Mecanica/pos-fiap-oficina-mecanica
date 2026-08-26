BEGIN;

CREATE TABLE usuario (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(254) NOT NULL UNIQUE,
    senha_hash VARCHAR(100) NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    criado_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE mecanico (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id UUID NOT NULL UNIQUE REFERENCES usuario (id),
    nome VARCHAR(200) NOT NULL
);

CREATE TABLE usuario_escopo (
    usuario_id UUID NOT NULL REFERENCES usuario (id),
    escopo VARCHAR(100) NOT NULL,
    PRIMARY KEY (usuario_id, escopo)
);

ALTER TABLE ordem_servico ADD COLUMN mecanico_responsavel_id UUID REFERENCES mecanico (id);
CREATE INDEX ix_ordem_servico_mecanico_responsavel_id ON ordem_servico (mecanico_responsavel_id);
ALTER TABLE auditoria_ordem_servico ADD CONSTRAINT fk_auditoria_usuario FOREIGN KEY (usuario_id) REFERENCES usuario (id);

COMMIT;
