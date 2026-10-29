ALTER TABLE problema_ordem_servico
    ADD COLUMN orcamento_id UUID REFERENCES orcamento (id),
    ADD COLUMN observacoes TEXT;

CREATE INDEX ix_problema_os_orcamento_id
    ON problema_ordem_servico (orcamento_id);
