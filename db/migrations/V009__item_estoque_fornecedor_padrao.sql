BEGIN;

ALTER TABLE item_estoque
    ADD COLUMN IF NOT EXISTS fornecedor_id UUID REFERENCES fornecedor (id);

CREATE INDEX IF NOT EXISTS ix_item_estoque_fornecedor_id
    ON item_estoque (fornecedor_id);

COMMIT;
