-- Метаданные файла для FileInfo (имя, MIME-тип, размер) без чтения диска на каждый ответ.
-- Новые загрузки заполняют обе колонки. У старых файлов тип восстанавливается
-- по расширению; размер остаётся NULL — сервер в этом случае берёт его с диска.

BEGIN;

ALTER TABLE file
    ADD COLUMN mime_type  varchar(100),
    ADD COLUMN size_bytes bigint CONSTRAINT file_size_valid CHECK (size_bytes >= 0);

UPDATE file SET mime_type = CASE lower(substring(storage_key from '\.([^./]+)$'))
    WHEN 'pdf'  THEN 'application/pdf'
    WHEN 'docx' THEN 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    WHEN 'pptx' THEN 'application/vnd.openxmlformats-officedocument.presentationml.presentation'
    WHEN 'xlsx' THEN 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    WHEN 'xls'  THEN 'application/vnd.ms-excel'
    WHEN 'zip'  THEN 'application/zip'
    WHEN 'rar'  THEN 'application/vnd.rar'
    WHEN 'jpg'  THEN 'image/jpeg'
    WHEN 'jpeg' THEN 'image/jpeg'
    WHEN 'png'  THEN 'image/png'
    WHEN 'webp' THEN 'image/webp'
    ELSE 'application/octet-stream' END
WHERE mime_type IS NULL;

COMMIT;
