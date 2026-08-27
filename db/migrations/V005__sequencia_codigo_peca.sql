BEGIN;

-- Codigo funcional da peca: PEC-000001, sequencia global, seis digitos, sem reset.
-- Mesmo padrao previsto para INS- (insumos) e SER- (servicos).
CREATE SEQUENCE IF NOT EXISTS seq_peca_codigo AS BIGINT START WITH 1 MINVALUE 1;

COMMIT;
