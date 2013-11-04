-- pgsql_061_content_storage.sql

BEGIN;

set search_path = im_storage, public;

-- 初始化 hash 表

CREATE OR REPLACE FUNCTION im_storage.hash_tables_init()
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
	tbname := 'im_storage.hash_' || suffix;
	
	IF NOT EXISTS(SELECT tablename FROM pg_catalog.pg_tables WHERE 
		schemaname = 'im_storage' AND tablename = tbname) THEN
	RAISE NOTICE 'tb is %', tbname;
	EXECUTE 'CREATE TABLE ' || tbname || '
	(
		LIKE im_storage.hash_template INCLUDING ALL , 
		CHECK (hashed LIKE ' || quote_literal(suffix||'%') || ')
	) 
	INHERITS (im_storage.hash_template)
	WITHOUT OIDS ;';
	count := count + 1;
	END IF;
END LOOP;

return count;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;

-- 初始化 mapping 表

CREATE OR REPLACE FUNCTION im_storage.mapping_tables_init()
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
	tbname := 'im_storage.mapping_' || suffix;

	IF NOT EXISTS(SELECT tablename FROM pg_catalog.pg_tables WHERE 
		schemaname = 'im_storage' AND tablename = tbname) THEN
	RAISE NOTICE 'tb is %', tbname;
	
	EXECUTE 'CREATE TABLE ' || tbname || '
	(
		LIKE im_storage.map_template INCLUDING ALL , 
		CHECK (id LIKE ' || quote_literal(suffix||'%') || ')
	) 
	INHERITS (im_storage.map_template)
	WITHOUT OIDS ;';
	count := count + 1;
	END IF;
END LOOP;

return count;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;


-- 保存 hash 记录
CREATE OR REPLACE FUNCTION im_storage.hash_save(a_hashed varchar, a_item_id varchar, a_path varchar, a_prefix varchar)

RETURNS int AS
$$
DECLARE
	suffix varchar;
	tbname varchar;
BEGIN

	suffix := substr(a_hashed, 1, 1);
	tbname := 'im_storage.hash_'||suffix;

	EXECUTE 'SELECT created FROM '||tbname||' WHERE hashed = '||quote_literal(a_hashed)|| ' LIMIT 1';

	IF FOUND THEN
		RETURN -1;
	END IF;

	EXECUTE 'INSERT INTO ' || tbname || '(hashed, item_id, prefix, path) VALUES (
		' || quote_literal(a_hashed) || ',' || quote_literal(a_item_id) || ',
		' || quote_literal(a_path) || ',' || quote_literal(a_prefix) || '
	)';

RETURN 1;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;


-- hash trigger
CREATE OR REPLACE FUNCTION hash_insert_trigger()
RETURNS TRIGGER AS $$
BEGIN
	PERFORM hash_save(NEW.hashed, NEW.item_id, NEW.path, NEW.prefix);
	RETURN NULL;
END;
$$
LANGUAGE plpgsql;

-- CREATE TRIGGER hash_insert_trigger BEFORE INSERT ON hash_template
-- FOR EACH ROW EXECUTE PROCEDURE hash_insert_trigger();


-- 保存 map 记录
CREATE OR REPLACE FUNCTION im_storage.map_save(a_id varchar, a_name varchar, a_path varchar, a_mime varchar, a_size, a_sev hstore)

RETURNS int AS
$$
DECLARE
	suffix varchar;
	tbname varchar;
BEGIN

	suffix := substr(a_id, 1, 1);
	tbname := 'im_storage.map_'||suffix;

	EXECUTE 'SELECT created FROM '||tbname||' WHERE id = '||quote_literal(a_id)|| ' LIMIT 1';

	IF FOUND THEN
		RETURN -1;
	END IF;

	EXECUTE 'INSERT INTO ' || tbname || '(id, name, path, mime, size, sev) VALUES (
		' || quote_literal(a_id) || ',' || quote_literal(a_name) || ',
		' || quote_literal(a_path) || ',' || quote_literal(a_mime) || ',
		' || a_size || ',' || quote_literal(a_sev) || '
	)';

RETURN 1;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;


-- 保存某条完整 entry 信息
CREATE OR REPLACE FUNCTION im_storage.entry_save
(a_id varchar, a_name varchar, a_path varchar, a_mime varchar, a_size, a_sev hstore
	, a_hashes []varchar, a_ids []varchar)

RETURNS int AS
$$
DECLARE
	suffix varchar;
	tb_hash varchar;
	tb_map varchar;
BEGIN

	-- TODO:

RETURN 1;
END;
$$
LANGUAGE 'plpgsql' VOLATILE;



END;


