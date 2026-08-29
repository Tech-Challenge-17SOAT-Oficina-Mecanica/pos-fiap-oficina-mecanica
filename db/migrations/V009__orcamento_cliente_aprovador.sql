ALTER TABLE orcamento
    ADD COLUMN IF NOT EXISTS cliente_aprovador_id UUID REFERENCES cliente (id);
