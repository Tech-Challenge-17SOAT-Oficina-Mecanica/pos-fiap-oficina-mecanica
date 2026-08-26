BEGIN;

ALTER TABLE servico
    DROP CONSTRAINT IF EXISTS fk_servico_usuario_desativacao,
    ADD CONSTRAINT fk_servico_usuario_desativacao
        FOREIGN KEY (usuario_desativacao) REFERENCES usuario (id);

COMMIT;
