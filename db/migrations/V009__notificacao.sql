BEGIN;

-- Fila de notificacoes ao cliente (DT-53). A operacao que dispara grava a linha como
-- PENDENTE na sua propria transacao e responde; o envio acontece depois, em processo
-- separado. Assim uma falha de e-mail nunca desfaz a operacao de negocio (RNF-OS-44),
-- e o resultado do envio fica registrado para permitir reenvio (RF-OS-88, RNF-OS-45).
CREATE TABLE IF NOT EXISTS notificacao (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cliente_id UUID NOT NULL REFERENCES cliente (id),
    canal VARCHAR(10) NOT NULL,
    tipo_evento VARCHAR(50) NOT NULL,
    -- Agregado que originou a notificacao, para rastrear a origem sem criar uma FK por
    -- contexto. Guarda o nome e o id, no mesmo formato da auditoria_ordem_servico.
    agregado VARCHAR(50) NOT NULL,
    agregado_id UUID NOT NULL,
    destinatario VARCHAR(254) NOT NULL,
    assunto VARCHAR(200) NOT NULL,
    conteudo TEXT NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'PENDENTE',
    tentativas INTEGER NOT NULL DEFAULT 0,
    ultimo_erro TEXT,
    criada_em TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    enviada_em TIMESTAMPTZ,
    CONSTRAINT ck_notificacao_canal CHECK (canal IN ('EMAIL')),
    CONSTRAINT ck_notificacao_status CHECK (status IN ('PENDENTE', 'ENVIADA', 'FALHOU')),
    CONSTRAINT ck_notificacao_tentativas CHECK (tentativas >= 0),
    CONSTRAINT ck_notificacao_enviada CHECK (
        (status = 'ENVIADA' AND enviada_em IS NOT NULL) OR
        (status <> 'ENVIADA' AND enviada_em IS NULL)
    )
);

-- O consumo da fila busca pendentes na ordem de criacao; indice parcial mantem o custo
-- proporcional ao que falta enviar, nao ao historico inteiro.
CREATE INDEX IF NOT EXISTS ix_notificacao_pendente
    ON notificacao (criada_em)
    WHERE status = 'PENDENTE';

CREATE INDEX IF NOT EXISTS ix_notificacao_agregado
    ON notificacao (agregado, agregado_id);

COMMIT;
