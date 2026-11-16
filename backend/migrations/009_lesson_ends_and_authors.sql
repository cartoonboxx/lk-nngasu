-- 1. Расписание: время окончания занятия хранится в БД (схема Lesson: ends_at,
--    последняя пара короче остальных — 80 минут). Существующим занятиям
--    конец проставляется как начало + 90 минут (LESSON_MINUTES по умолчанию);
--    если у вас другая длительность пары, поправьте интервал ниже.
-- 2. Авторы материалов и заданий — пользователи, а не только преподаватели:
--    по контракту темы, материалы и задания может создавать и администратор.
-- 3. Название задания — до 255 символов.

BEGIN;

ALTER TABLE lesson ADD COLUMN ends_at time;
UPDATE lesson SET ends_at = starts_at + interval '90 minutes';
ALTER TABLE lesson
    ALTER COLUMN ends_at SET NOT NULL,
    ADD CONSTRAINT lesson_interval_valid CHECK (ends_at > starts_at);

ALTER TABLE material   DROP CONSTRAINT material_created_by_fkey,
    ADD CONSTRAINT material_created_by_fkey   FOREIGN KEY (created_by) REFERENCES app_user (id);
ALTER TABLE assignment DROP CONSTRAINT assignment_created_by_fkey,
    ADD CONSTRAINT assignment_created_by_fkey FOREIGN KEY (created_by) REFERENCES app_user (id);

ALTER TABLE assignment ALTER COLUMN title TYPE varchar(255);

COMMIT;
