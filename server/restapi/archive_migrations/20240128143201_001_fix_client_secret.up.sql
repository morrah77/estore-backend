alter table payments drop column client_secret;
alter table payments add column checkout_session_id text(255) default '';
alter table payments add column 'payment_intent_id' text(255) default '';
