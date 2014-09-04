-- pgsql_11_imsto_auth.sql
-- 验证相关

BEGIN;

SET search_path = imsto, public;

CREATE OR REPLACE FUNCTION imsto.app_save(a_name varchar, a_api_key varchar)
RETURNS int AS
$$
DECLARE t_id int;

BEGIN

	SELECT id INTO t_id FROM apps WHERE api_key = a_api_key;
	IF t_id IS NOT NULL THEN
		RETURN t_id;
	END IF;

	INSERT INTO apps(name, api_key) VALUES (a_name, a_api_key) RETURNING id INTO t_id;
	IF t_id IS NOT NULL THEN
		RETURN t_id;
	END IF;

RETURN 0;

END;
$$
LANGUAGE 'plpgsql' VOLATILE;



-- 更新 ticket
CREATE OR REPLACE FUNCTION imsto.ticket_update(a_id int, a_item_id varchar)
RETURNS int AS
$$
DECLARE
	t_roof varchar;
	tb_meta varchar;
	t_path varchar;

BEGIN

	SELECT roof INTO t_roof FROM upload_ticket WHERE id = a_id;
	IF t_roof IS NULL THEN
		RETURN -1;
	END IF;

	tb_meta := 'meta_' || t_roof;

	EXECUTE 'SELECT path FROM '||tb_meta||' WHERE id = $1 LIMIT 1'
	INTO t_path
	USING a_item_id;

	IF t_path IS NULL THEN
		RAISE NOTICE 'item not found %', a_item_id;
		RETURN -2;
	END IF;

	UPDATE upload_ticket
	SET img_id=a_item_id, img_path=t_path, done=true, updated=CURRENT_TIMESTAMP
	WHERE id = a_id;

RETURN 1;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;



END;
