BEGIN;

-- Codigo funcional do insumo: INS-000001, sequencia global, seis digitos, sem reset.
-- Mesmo padrao do PEC- em pecas e do SER- previsto para servicos.
CREATE SEQUENCE IF NOT EXISTS seq_insumo_codigo AS BIGINT START WITH 1 MINVALUE 1;

COMMIT;
