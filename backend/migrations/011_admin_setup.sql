-- Право владельца отделено от доступа к обычной административной панели.
ALTER TABLE app_user ADD COLUMN can_manage_admins boolean NOT NULL DEFAULT false;
ALTER TABLE app_user ADD CONSTRAINT owner_is_admin CHECK (NOT can_manage_admins OR is_admin);
CREATE UNIQUE INDEX app_user_single_owner ON app_user ((true)) WHERE can_manage_admins;

-- При обновлении существующей установки сохраняем доступ к управлению.
UPDATE app_user SET can_manage_admins = true
WHERE id = (SELECT id FROM app_user WHERE is_admin ORDER BY is_active DESC, id LIMIT 1);

CREATE TABLE admin_setup (
    id boolean PRIMARY KEY DEFAULT true CHECK (id),
    completed_at timestamptz
);
INSERT INTO admin_setup (id, completed_at)
VALUES (true, CASE WHEN EXISTS (SELECT 1 FROM app_user WHERE is_admin) THEN now() END);

-- Любое создание администратора, включая старую CLI-команду, закрывает настройку.
-- Удаление/блокировка пользователей и смена ключа не открывают её обратно.
CREATE FUNCTION close_admin_setup() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.is_admin THEN
        UPDATE admin_setup SET completed_at = coalesce(completed_at, now()) WHERE id;
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER app_user_close_setup AFTER INSERT OR UPDATE OF is_admin ON app_user
FOR EACH ROW EXECUTE FUNCTION close_admin_setup();

ALTER TABLE user_invitation ADD COLUMN grants_admin boolean NOT NULL DEFAULT false;

CREATE TABLE admin_access_audit (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor_id int REFERENCES app_user (id) ON DELETE SET NULL,
    target_id int REFERENCES app_user (id) ON DELETE SET NULL,
    target_email varchar(255) NOT NULL,
    action varchar(32) NOT NULL CHECK (action IN ('setup', 'invite', 'promote', 'activate', 'cli_create')),
    created_at timestamptz NOT NULL DEFAULT now()
);
