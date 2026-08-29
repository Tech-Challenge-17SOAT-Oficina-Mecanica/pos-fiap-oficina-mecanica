ALTER TABLE ordem_servico ADD COLUMN data_entrada_fila TIMESTAMPTZ;

CREATE INDEX ix_ordem_servico_fila_atendimento
    ON ordem_servico ((mecanico_responsavel_id IS NULL), data_entrada_fila, id)
    WHERE status = 'AGUARDANDO_EXECUCAO' AND data_entrada_fila IS NOT NULL;
