BEGIN;

ALTER TABLE cliente ADD COLUMN motivo_inativacao VARCHAR(200);

COMMIT;
