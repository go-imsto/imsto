-- pgsql_061_content_storage.sql

BEGIN;

set search_path = imsto, public;

-- 初始化 hash 表

CREATE OR REPLACE FUNCTION imsto.hash_tables_init()
RETURNS int AS
$$
DECLARE
	count int;
	suffix varchar;
	tbname varchar;
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

return count;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;

-- 初始化 mapping 表

CREATE OR REPLACE FUNCTION imsto.mapping_tables_init()
RETURNS int AS
$$
DECLARE
	basestr varchar;
	count int;
	suffix varchar;
	tbname varchar;
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

return count;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;


-- 保存 hash 记录
CREATE OR REPLACE FUNCTION imsto.hash_save(a_hashed varchar, a_item_id varchar, a_path varchar)

RETURNS int AS
$$
DECLARE
	suffix varchar;
	tbname varchar;
	t_id varchar;
BEGIN

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

-- check hash exists
-- CREATE OR REPLACE FUNCTION imsto.hash_exists(a_hashed varchar)

-- 保存 map 记录
CREATE OR REPLACE FUNCTION imsto.map_save(
	a_id varchar, a_path varchar, a_name varchar, a_mime varchar, a_size int, a_sev hstore)

RETURNS int AS
$$
DECLARE
	suffix varchar;
	tbname varchar;
	t_st smallint;
BEGIN

	suffix := substr(a_id, 1, 1);
	tbname := 'imsto.mapping_'||suffix;

	EXECUTE 'SELECT status FROM '||tbname||' WHERE id = $1 LIMIT 1'
	INTO t_st
	USING a_id;

	IF t_st IS NOT NULL THEN
		RAISE NOTICE 'exists map %', t_st;
		RETURN -1;
	END IF;

	EXECUTE 'INSERT INTO ' || tbname || '(id, path, name, mime, size, sev) VALUES (
		$1, $2, $3, $4, $5, $6
	)'
	USING a_id, a_path, a_name, a_mime, a_size, a_sev;

RETURN 1;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;


-- 保存某条完整 entry 信息
CREATE OR REPLACE FUNCTION imsto.entry_save (a_section varchar,
	a_id varchar, a_path varchar, a_name varchar, a_mime varchar, a_size int
	, a_meta hstore, a_sev hstore, a_hashes varchar[], a_ids varchar[])

RETURNS int AS
$$
DECLARE
	m_v varchar;
	tb_hash varchar;
	tb_map varchar;
	tb_meta varchar;
	t_path varchar;
BEGIN

	tb_meta := 'meta_' || a_section;

	EXECUTE 'SELECT path FROM '||tb_meta||' WHERE id = $1 LIMIT 1'
	INTO t_path
	USING a_id;

	IF t_path IS NOT NULL THEN
		RAISE NOTICE 'exists meta %', t_path;
		RETURN -1;
	END IF;

	-- save entry hashes
	FOR m_v IN SELECT UNNEST(a_hashes) AS value LOOP
		PERFORM hash_save(m_v, a_id, a_path);
	END LOOP;

	-- save entry map
	FOR m_v IN SELECT UNNEST(a_ids) AS value LOOP
		PERFORM map_save(m_v, a_path, a_name, a_mime, a_size, a_sev);
	END LOOP;

	-- save entry meta
	EXECUTE 'INSERT INTO ' || tb_meta || '(id, path, name, meta, hashes, ids, size, sev) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8
	)'
	USING a_id, a_path, a_name, a_meta, a_hashes, a_ids, a_size, a_sev;

RETURN 1;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;



END;


