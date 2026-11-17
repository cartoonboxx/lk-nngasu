-- Очередь исходящих писем (приглашения задать пароль).
-- Письмо пишется в той же транзакции, что и создание пользователя, а отправляет
-- его фоновый обработчик с повторными попытками. После отправки тело стирается:
-- в нём одноразовая ссылка активации.

CREATE TABLE mail_outbox (
    id              bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    recipient       varchar(255) NOT NULL,
    subject         varchar(255) NOT NULL,
    body            text         NOT NULL,
    attempts        int          NOT NULL DEFAULT 0,
    next_attempt_at timestamptz  NOT NULL DEFAULT now(),
    last_error      text,
    sent_at         timestamptz,
    failed_at       timestamptz,  -- попытки исчерпаны
    created_at      timestamptz  NOT NULL DEFAULT now()
);

-- Обработчик выбирает только неотправленные письма, чей срок попытки наступил.
CREATE INDEX mail_outbox_pending_idx ON mail_outbox (next_attempt_at)
    WHERE sent_at IS NULL AND failed_at IS NULL;
