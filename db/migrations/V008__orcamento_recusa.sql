BEGIN;

ALTER TABLE orcamento ADD COLUMN cliente_recusador_id UUID REFERENCES cliente (id);
ALTER TABLE orcamento ADD COLUMN motivo_recusa VARCHAR(500);

COMMIT;
