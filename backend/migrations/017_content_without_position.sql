-- Темы и материалы по контракту 1.2 (scheme.dbml): ручного порядка нет — сортировка по id;
-- у материала нет названия, автора и даты создания (поля закомментированы в схеме).
-- Материал — это файл и/или внешняя ссылка.

ALTER TABLE topic DROP COLUMN position;
ALTER TABLE material
    DROP COLUMN position,
    DROP COLUMN title,
    DROP COLUMN created_by,
    DROP COLUMN created_at;
