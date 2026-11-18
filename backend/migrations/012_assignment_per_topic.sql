-- Одна тема — не более одного задания (контракт: связь assignment.topic_id — topic.id
-- один к одному; повтор — 409 topic_assignment_exists).
--
-- Если в какой-то теме уже несколько заданий, миграция остановится и назовёт тему:
-- перенесите лишние задания в другие темы (PUT /assignments/:id) и перезапустите сервер.

DO $$
DECLARE dup int;
BEGIN
    SELECT topic_id INTO dup FROM assignment GROUP BY topic_id HAVING count(*) > 1 LIMIT 1;
    IF dup IS NOT NULL THEN
        RAISE EXCEPTION 'migration 012: topic % has more than one assignment; move extra assignments to other topics and restart', dup;
    END IF;
END $$;

ALTER TABLE assignment ADD CONSTRAINT assignment_topic_unique UNIQUE (topic_id);
