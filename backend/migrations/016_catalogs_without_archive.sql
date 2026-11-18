-- Справочники по контракту 1.2: архивного состояния нет (поля archived в scheme.dbml нет),
-- удаление — только записи без ссылок. Записи, бывшие в архиве, снова видны в списках.
--
-- Заодно — остальные поля справочников и нагрузки по scheme.dbml и схемам API:
--   * группа уникальна парой «название + год набора» (без учёта регистра);
--   * год набора и год начала семестра — от 2000 (smallint ограничивает сверху);
--   * название семестра вычисляется из года и номера, а не вводится;
--   * planned_hours у нагрузки может быть null, точность — два знака.

ALTER TABLE department     DROP COLUMN archived;
ALTER TABLE subject        DROP COLUMN archived;
ALTER TABLE specialization DROP COLUMN archived;
ALTER TABLE study_group    DROP COLUMN archived;
ALTER TABLE semester       DROP COLUMN archived;
ALTER TABLE room           DROP COLUMN archived;

DROP INDEX study_group_name_ci_idx;
CREATE UNIQUE INDEX study_group_name_year_ci_idx ON study_group (lower(name), admission_year);
ALTER TABLE study_group
    DROP CONSTRAINT study_group_admission_year_valid,
    ADD CONSTRAINT study_group_admission_year_valid CHECK (admission_year >= 2000);

ALTER TABLE semester
    DROP CONSTRAINT semester_year_valid,
    ADD CONSTRAINT semester_year_valid CHECK (academic_year_start >= 2000),
    DROP COLUMN name;
ALTER TABLE semester ADD COLUMN name text GENERATED ALWAYS AS (
    academic_year_start::text || '/' || lpad(((academic_year_start + 1) % 100)::text, 2, '0')
    || ' · ' || CASE number WHEN 1 THEN 'Осенний' ELSE 'Весенний' END
) STORED;

ALTER TABLE workload
    ALTER COLUMN planned_hours DROP NOT NULL,
    DROP CONSTRAINT workload_hours_step;
