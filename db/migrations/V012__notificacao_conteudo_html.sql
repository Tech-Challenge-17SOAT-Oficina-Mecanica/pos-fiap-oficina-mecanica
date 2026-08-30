BEGIN;

-- Corpo alternativo em HTML. O orcamento enviado ao cliente traz uma tabela de itens,
-- que em texto puro perde o alinhamento na maioria dos clientes de e-mail.
-- Nulo significa que a notificacao tem apenas a versao em texto.
ALTER TABLE notificacao
    ADD COLUMN IF NOT EXISTS conteudo_html TEXT;

COMMIT;
