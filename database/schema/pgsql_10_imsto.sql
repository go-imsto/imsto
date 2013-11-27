-- pgsql_061_content_storage.sql

-- CREATE EXTENSION hstore;

BEGIN;

CREATE SCHEMA imsto;
COMMENT ON SCHEMA imsto IS '存储相关';
GRANT ALL ON SCHEMA imsto TO imsto;

set search_path = imsto, public;

CREATE DOMAIN entry_id AS TEXT
CHECK(
	VALUE ~ '^[a-z0-9]{20,36}$'
);

CREATE DOMAIN entry_path AS TEXT
CHECK (
	VALUE ~ '^[a-z0-9]{2}/[a-z0-9]{2}/[a-z0-9]{16,32}\.[a-z0-9]{2,6}$'
);


CREATE SEQUENCE hash_id_seq;

-- all file hash values
CREATE TABLE hash_template (
	id bigint DEFAULT nextval('hash_id_seq'),
	hashed varCHAR(40) NOT NULL UNIQUE , 
	item_id entry_id NOT NULL , 
	-- prefix varCHAR(10) NOT NULL DEFAULT '' , 
	path entry_path NOT NULL , 
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) WITHOUT OIDS;

-- mapping for id and storage engine item
CREATE TABLE map_template (
	id entry_id NOT NULL, -- id = base_convert(hash,16,36)
	name name NOT NULL DEFAULT '',
	path entry_path NOT NULL , 
	mime varCHAR(64) NOT NULL DEFAULT '' , 
	size int NOT NULL DEFAULT 0 CHECK (size >= 0),
	sev hstore, -- storage info
	status smallint NOT NULL DEFAULT 0, -- 0=valid,1=deleted
	roofs varCHAR(12)[] NOT NULL DEFAULT '{}',
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) WITHOUT OIDS;

-- meta browsable
CREATE TABLE meta_template (
	id entry_id NOT NULL,
	path entry_path NOT NULL , 
	name name NOT NULL DEFAULT '',
	meta hstore,
	hashes varCHAR(40)[],
	ids varCHAR(38)[],
	-- mime varCHAR(64) NOT NULL DEFAULT '' , 
	size int NOT NULL DEFAULT 0,
	sev hstore,
	exif hstore, -- exif info
	app_id int NOT NULL DEFAULT 0,
	author int NOT NULL DEFAULT 0,
	status smallint NOT NULL DEFAULT 0, -- 0=valid,1=deleted
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) WITHOUT OIDS;
CREATE INDEX idx_meta_created ON meta_template (status, created) ;

CREATE TABLE meta_common
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

CREATE TABLE meta_demo
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

CREATE TABLE meta_wpitem
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

CREATE TABLE meta_crafts
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

CREATE TABLE meta_avatar
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

CREATE TABLE meta_wpcms
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;


END;
