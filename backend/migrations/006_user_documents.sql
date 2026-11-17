-- Паспорт и СНИЛС пользователя (AdminUser в документации API).
-- Хранятся только нормализованные цифры. Колонки допускают NULL: у пользователей,
-- заведённых раньше, документов нет; новые пользователи создаются с ними всегда.

ALTER TABLE app_user
    ADD COLUMN passport varchar(10)
        CONSTRAINT app_user_passport_digits CHECK (passport ~ '^[0-9]{10}$'),
    ADD COLUMN snils    varchar(11)
        CONSTRAINT app_user_snils_digits CHECK (snils ~ '^[0-9]{11}$');

-- Переносим документы из уже одобренных заявок.
UPDATE app_user u
SET passport = r.passport, snils = r.snils
FROM registration_request r
WHERE r.user_id = u.id AND r.status = 'approved' AND u.passport IS NULL;
