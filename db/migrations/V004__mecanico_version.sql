BEGIN;

ALTER TABLE mecanico ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE mecanico ADD CONSTRAINT ck_mecanico_version CHECK (version > 0);

COMMIT;
