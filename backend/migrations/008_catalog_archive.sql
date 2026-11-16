-- Справочники по схемам NamedCatalog, Specialization, Group, Semester, Room:
-- признак archived, названия до 200 символов, уникальность без учёта регистра,
-- название семестра; нагрузка — planned_hours обязательны, шаг 0.25.
--
-- Перед запуском проверь дубликаты, отличающиеся только регистром,
-- иначе создание уникальных индексов упадёт:
--   SELECT lower(name), count(*) FROM department  GROUP BY 1 HAVING count(*) > 1;
--   SELECT lower(name), count(*) FROM subject     GROUP BY 1 HAVING count(*) > 1;
--   SELECT lower(name), count(*) FROM study_group GROUP BY 1 HAVING count(*) > 1;
--   SELECT lower(number), count(*) FROM room      GROUP BY 1 HAVING count(*) > 1;

BEGIN;

-- Кафедры и дисциплины (NamedCatalog).
ALTER TABLE department
    ADD COLUMN archived bool NOT NULL DEFAULT false,
    ALTER COLUMN name TYPE varchar(200);
CREATE UNIQUE INDEX department_name_ci_idx ON department (lower(name));

ALTER TABLE subject ADD COLUMN archived bool NOT NULL DEFAULT false;
CREATE UNIQUE INDEX subject_name_ci_idx ON subject (lower(name));

-- Направления.
ALTER TABLE specialization ADD COLUMN archived bool NOT NULL DEFAULT false;

-- Группы: название уникально само по себе (а не в паре с годом набора), год 2000–2100.
ALTER TABLE study_group
    ADD COLUMN archived bool NOT NULL DEFAULT false,
    ALTER COLUMN name TYPE varchar(200),
    DROP CONSTRAINT IF EXISTS study_group_name_admission_year_key,
    DROP CONSTRAINT study_group_admission_year_valid,
    ADD CONSTRAINT study_group_admission_year_valid CHECK (admission_year BETWEEN 2000 AND 2100);
CREATE UNIQUE INDEX study_group_name_ci_idx ON study_group (lower(name));

-- Семестры: название для отображения («2026/27 · Осенний»), год 2000–2100.
ALTER TABLE semester
    ADD COLUMN name varchar(200),
    ADD COLUMN archived bool NOT NULL DEFAULT false,
    ADD CONSTRAINT semester_year_valid CHECK (academic_year_start BETWEEN 2000 AND 2100);
UPDATE semester
SET name = academic_year_start || '/' || lpad(((academic_year_start + 1) % 100)::text, 2, '0')
           || ' · ' || CASE number WHEN 1 THEN 'Осенний' ELSE 'Весенний' END;
ALTER TABLE semester ALTER COLUMN name SET NOT NULL;

-- Аудитории: обозначение (колонка number, наружу — name) до 200 символов.
ALTER TABLE room
    ADD COLUMN archived bool NOT NULL DEFAULT false,
    ALTER COLUMN number TYPE varchar(200);
CREATE UNIQUE INDEX room_number_ci_idx ON room (lower(number));

-- Нагрузка: planned_hours обязательны, неотрицательны, шаг 0.25 академического часа.
UPDATE workload SET planned_hours = 0 WHERE planned_hours IS NULL;
ALTER TABLE workload
    ALTER COLUMN planned_hours SET NOT NULL,
    ADD CONSTRAINT workload_hours_step CHECK (planned_hours * 4 = trunc(planned_hours * 4));

COMMIT;
