-- ФИО в заявке на регистрацию.
-- В форме регистрации ФИО нет: его вводит администратор при одобрении,
-- и оно сохраняется в заявке — по нему работает поиск в списке заявок (q).
-- У необработанной и отклонённой заявки ФИО — NULL: неизвестное не заменяется вымышленным.
ALTER TABLE registration_request
    ADD COLUMN IF NOT EXISTS last_name    varchar(50),
    ADD COLUMN IF NOT EXISTS first_name   varchar(50),
    ADD COLUMN IF NOT EXISTS middle_names varchar(50);
