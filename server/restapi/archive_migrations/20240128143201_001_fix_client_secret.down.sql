alter table payments drop column checkout_session_id;
alter table payments drop column payment_intent_id;
alter table payments add column client_secret text(255) default '';
