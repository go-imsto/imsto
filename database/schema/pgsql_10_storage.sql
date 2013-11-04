-- pgsql_061_content_storage.sql

-- CREATE EXTENSION hstore;

BEGIN;

CREATE SCHEMA im_storage;
COMMENT ON SCHEMA im_storage IS '存储相关';
GRANT ALL ON SCHEMA im_storage TO wp_content;

set search_path = im_storage, public;

CREATE SEQUENCE hash_id_seq;

-- all file hash values
CREATE TABLE hash_template (
	id bigint DEFAULT nextval('hash_id_seq'),
	hashed varCHAR(40) NOT NULL UNIQUE , 
	item_id varCHAR(36) NOT NULL DEFAULT '' , 
	prefix varCHAR(10) NOT NULL DEFAULT '' , 
	path varCHAR(39) NOT NULL DEFAULT '' , 
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) WITHOUT OIDS;

-- mapping for id and storage engine item
CREATE TABLE map_template (
	id varCHAR(38) NOT NULL, -- id = base_convert(hash,16,36)
	name name NOT NULL DEFAULT '',
	path varCHAR(39) NOT NULL DEFAULT '' , 
	mime varCHAR(64) NOT NULL DEFAULT '' , 
	size int NOT NULL DEFAULT 0,
	sev hstore, -- storage info
	status smallint NOT NULL DEFAULT 0, -- 0=valid,1=deleted
	PRIMARY KEY (id)
) WITHOUT OIDS;

-- meta browsable
CREATE TABLE meta_template (
	id varCHAR(38) NOT NULL,
	hashes varCHAR(40)[],
	ids varCHAR(38)[],
	name name NOT NULL DEFAULT '',
	path varCHAR(39) NOT NULL DEFAULT '' , 
	mime varCHAR(64) NOT NULL DEFAULT '' , 
	size int NOT NULL DEFAULT 0,
	meta hstore,
	sev hstore, -- storage info
	app_id int NOT NULL DEFAULT 0,
	author int NOT NULL DEFAULT 0,
	status smallint NOT NULL DEFAULT 0, -- 0=valid,1=deleted
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) WITHOUT OIDS;
CREATE INDEX idx_meta_created ON meta_template (created) ;

CREATE TABLE meta_wpitem
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;


END;
