BEGIN;

-- Indice que sustenta a consulta da fila de atendimento. A coluna data_entrada_fila
-- e criada na V010; aqui so entra o indice, para nao duplicar o ALTER TABLE.
CREATE INDEX IF NOT EXISTS ix_ordem_servico_fila_atendimento
    ON ordem_servico ((mecanico_responsavel_id IS NULL), data_entrada_fila, id)
    WHERE status = 'AGUARDANDO_EXECUCAO' AND data_entrada_fila IS NOT NULL;

COMMIT;
