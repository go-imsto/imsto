-- pgsql_061_content_storage.sql

BEGIN;

set search_path = imsto, public;

-- 初始化 hash 表

CREATE OR REPLACE FUNCTION imsto.hash_tables_init()
RETURNS int AS
$$
DECLARE
	count int;
	suffix text;
	tbname text;
BEGIN

	count := 0;
-- 创建 表
FOR i IN 0..15 LOOP
	-- some computations here
	suffix := to_hex(i%16);
	tbname := 'hash_' || suffix;
	
	IF NOT EXISTS(SELECT tablename FROM pg_catalog.pg_tables WHERE 
		schemaname = 'imsto' AND tablename = tbname) THEN
	RAISE NOTICE 'tb is %', tbname;
	EXECUTE 'CREATE TABLE imsto.' || tbname || '
	(
		LIKE imsto.hash_template INCLUDING ALL , 
		CHECK (hashed LIKE ' || quote_literal(suffix||'%') || ')
	) 
	INHERITS (imsto.hash_template)
	WITHOUT OIDS ;';
	count := count + 1;
	END IF;
END LOOP;

RETURN count;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;

-- 初始化 mapping 表

CREATE OR REPLACE FUNCTION imsto.mapping_tables_init()
RETURNS int AS
$$
DECLARE
	basestr text;
	count int;
	suffix text;
	tbname text;
BEGIN

	count := 0;
	basestr := '0123456789abcdefghijklmnopqrstuvwxyz';
-- 创建 表
FOR i IN 1..36 LOOP
	-- some computations here
	suffix := substr(basestr, i, 1);
	tbname := 'mapping_' || suffix;

	IF NOT EXISTS(SELECT tablename FROM pg_catalog.pg_tables WHERE 
		schemaname = 'imsto' AND tablename = tbname) THEN
	RAISE NOTICE 'tb is %', tbname;
	
	EXECUTE 'CREATE TABLE imsto.' || tbname || '
	(
		LIKE imsto.map_template INCLUDING ALL , 
		CHECK (id LIKE ' || quote_literal(suffix||'%') || ')
	) 
	INHERITS (imsto.map_template)
	WITHOUT OIDS ;';
	count := count + 1;
	END IF;
END LOOP;

RETURN count;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;


-- 保存 hash 记录
CREATE OR REPLACE FUNCTION imsto.hash_save(a_hashed text, a_item_id text, a_path text)

RETURNS int AS
$$
DECLARE
	suffix text;
	tbname text;
	t_id text;
BEGIN

	IF char_length(a_hashed) < 20 THEN
		RAISE NOTICE 'bad hash value {%}', a_hashed;
		RETURN -1;
	END IF;

	suffix := substr(a_hashed, 1, 1);
	tbname := 'imsto.hash_'||suffix;

	EXECUTE 'SELECT item_id FROM '||tbname||' WHERE hashed = $1 LIMIT 1'
	INTO t_id
	USING a_hashed;

	IF t_id IS NOT NULL THEN
		RAISE NOTICE 'exists hash %(%)', a_hashed, t_id;
		RETURN -1;
	END IF;

	EXECUTE 'INSERT INTO '||tbname||' (hashed, item_id, path) VALUES (
		$1, $2, $3
	)'
	USING a_hashed, a_item_id, a_path;

RETURN 1;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;


-- hash trigger
CREATE OR REPLACE FUNCTION hash_insert_trigger()
RETURNS TRIGGER AS $$
BEGIN
	PERFORM hash_save(NEW.hashed, NEW.item_id, NEW.path);
	RETURN NULL;
END;
$$
LANGUAGE plpgsql;

-- CREATE TRIGGER hash_insert_trigger BEFORE INSERT ON hash_template
-- FOR EACH ROW EXECUTE PROCEDURE hash_insert_trigger();


-- 保存 map 记录
CREATE OR REPLACE FUNCTION imsto.map_save(
	a_id text, a_path text, a_name text, a_mime text, a_size int, a_sev hstore, a_roof text)

RETURNS int AS
$$
DECLARE
	suffix text;
	tbname text;
	t_st smallint;
	t_roofs text[];
	i_roofs text[];
BEGIN

	suffix := substr(a_id, 1, 1);
	tbname := 'imsto.mapping_'||suffix;
	i_roofs := ('{' || a_roof || '}')::text[];

	EXECUTE 'SELECT roofs FROM '||tbname||' WHERE id = $1 LIMIT 1'
	INTO t_roofs
	USING a_id;

	IF t_roofs IS NOT NULL THEN
		RAISE NOTICE 'exists map %', t_roofs;
		-- TODO: merge roofs
		IF NOT t_roofs @> i_roofs THEN
		t_roofs := t_roofs || i_roofs;
		EXECUTE 'UPDATE ' || tbname || ' SET roofs = $1 WHERE id = $2'
		USING t_roofs, a_id;
		END IF;
		RETURN -1;
	ELSE
		t_roofs := i_roofs;
	END IF;

	EXECUTE 'INSERT INTO ' || tbname || '(id, path, name, mime, size, sev, roofs) VALUES (
		$1, $2, $3, $4, $5, $6, $7
	)'
	USING a_id, a_path, a_name, a_mime, a_size, a_sev, t_roofs;

RETURN 1;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;


-- FUNCTION: entry_save(text, text, text, hstore, hstore, text[], text[], smallint, integer)

-- DROP FUNCTION entry_save(text, text, text, hstore, hstore, text[], text[], smallint, integer);

-- 保存某条完整 entry 信息
CREATE OR REPLACE FUNCTION imsto.entry_save (a_roof text,
	a_id text, a_path text, a_meta hstore, a_sev hstore
	, a_hashes text[], a_ids text[]
	, a_appid int, a_author int)

RETURNS int AS
$$
DECLARE
	m_v text;
	tb_hash text;
	tb_map text;
	tb_meta text;
	t_status smallint;
BEGIN

	tb_meta := 'meta_' || a_roof;

	EXECUTE 'SELECT status FROM '||tb_meta||' WHERE id = $1 LIMIT 1'
	INTO t_status
	USING a_id;

	IF t_status IS NOT NULL THEN
		RAISE NOTICE 'exists meta %', t_status;
		IF t_status = 1 THEN -- deleted, so restore it
			EXECUTE 'UPDATE ' || tb_meta || ' SET status = 0 WHERE id = $1'
			USING a_id;
			RETURN -2;
		END IF;
		RETURN -1;
	END IF;

	-- save entry hashes
	FOR m_v IN SELECT UNNEST(a_hashes) AS value LOOP
		PERFORM hash_save(m_v, a_id, a_path);
	END LOOP;

	-- save entry map
	FOR m_v IN SELECT UNNEST(a_ids) AS value LOOP
		PERFORM map_save(m_v, a_path, a_meta->'name', a_meta->'mime', (a_meta->'size')::int, a_sev, a_roof);
	END LOOP;

	IF NOT a_ids @> ARRAY[a_id] THEN
		PERFORM map_save(a_id, a_path, a_meta->'name', a_meta->'mime', (a_meta->'size')::int, a_sev, a_roof);
	END IF;

	-- save entry meta
	EXECUTE 'INSERT INTO ' || tb_meta || '(id, path, name, size, meta, hashes, ids, sev, app_id, author, roof)
	 VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
	)'
	USING a_id, a_path, a_meta->'name', (a_meta->'size')::int, a_meta, a_hashes, a_ids, a_sev, a_appid, a_author, a_roof;

RETURN 1;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;


-- 预先保存某条完整 entry 信息
CREATE OR REPLACE FUNCTION imsto.entry_ready (a_roof text,
	a_id text, a_path text, a_meta hstore
	, a_hashes text[], a_ids text[]
	, a_appid smallint, a_author int)

RETURNS int AS
$$
BEGIN

IF NOT EXISTS(SELECT created FROM meta__prepared WHERE id = a_id) THEN
	INSERT INTO meta__prepared (id, roof, path, meta, hashes, ids, app_id, author)
	VALUES (a_id, a_roof, a_path, a_meta, a_hashes, a_ids, a_appid, a_author);
	IF FOUND THEN
		RETURN 1;
	ELSE
		RETURN -1;
	END IF;
ELSE
	RETURN -2;
END IF;

END;
$$
LANGUAGE 'plpgsql' VOLATILE;

CREATE OR REPLACE FUNCTION imsto.entry_set_done(a_id text, a_sev hstore)
RETURNS int AS
$$
DECLARE
	m_rec RECORD;
	t_ret int;
BEGIN

SELECT * FROM meta__prepared WHERE id = a_id INTO m_rec;
IF NOT FOUND THEN
	RETURN -2;
END IF;

SELECT entry_save(m_rec.roof, m_rec.id, m_rec.path, m_rec.meta, a_sev,
 m_rec.hashes, m_rec.ids, m_rec.app_id, m_rec.author) INTO t_ret;

DELETE FROM meta__prepared WHERE id = a_id;

RETURN t_ret;

END;
$$
LANGUAGE 'plpgsql' VOLATILE;


CREATE OR REPLACE FUNCTION imsto.entry_delete(a_roof text, a_id text)
RETURNS int AS
$$
DECLARE
	tb_meta text;
	t_status smallint;
BEGIN
	tb_meta := 'meta_' || a_roof;

	EXECUTE 'SELECT status FROM '||tb_meta||' WHERE id = $1 LIMIT 1'
	INTO t_status
	USING a_id;

	IF t_status IS NULL THEN
		RETURN -1;
	END IF;

	IF NOT EXISTS (SELECT status FROM meta__deleted WHERE id = a_id) THEN
		EXECUTE 'INSERT INTO meta__deleted SELECT *, $1, CURRENT_TIMESTAMP FROM '||tb_meta||' WHERE id = $2'
		USING a_roof, a_id;
	END IF;

	EXECUTE 'DELETE FROM '||tb_meta||' WHERE id = $1'
	USING a_id;

	-- TODO: delete mapping and hash

	RETURN 1;

END;
$$
LANGUAGE 'plpgsql' VOLATILE;



END;

/*
SET search_path = imsto;
SELECT hash_tables_init();
SELECT mapping_tables_init();
*/


