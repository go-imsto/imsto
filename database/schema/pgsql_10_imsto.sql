-- pgsql_10_imsto.sql

-- CREATE EXTENSION hstore;
-- require: >= postgresql-9.3;

BEGIN;

CREATE SCHEMA imsto;
COMMENT ON SCHEMA imsto IS '存储相关';
GRANT ALL ON SCHEMA imsto TO imsto;

set search_path = imsto, public;

CREATE DOMAIN entry_id AS TEXT
CHECK(
	VALUE ~ '^[a-z0-9]{9,36}$'
);

CREATE DOMAIN entry_path AS TEXT
CHECK (
	VALUE ~ '^[a-z0-9]{2}/?[a-z0-9]{2}/?[a-z0-9]{5,32}\.[a-z0-9]{2,6}$'
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
	id entry_id NOT NULL, -- id = base_convert(hash,16,36), hash = crc64Sum
	name name NOT NULL DEFAULT '',
	path entry_path NOT NULL ,
	mime varCHAR(64) NOT NULL DEFAULT '' ,
	size int NOT NULL DEFAULT 0 CHECK (size >= 0),
	sev hstore NOT NULL DEFAULT '', -- storage info
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
	roof varCHAR(12) NOT NULL DEFAULT '',
	meta hstore NOT NULL DEFAULT '',
	hashes varCHAR(40)[],
	ids varCHAR(38)[],
	size int NOT NULL DEFAULT 0,
	sev hstore NOT NULL DEFAULT '',
	exif hstore NOT NULL DEFAULT '', -- exif info
	app_id smallint NOT NULL DEFAULT 0,
	author int NOT NULL DEFAULT 0,
	status smallint NOT NULL DEFAULT 0, -- 0=valid,1=hidden
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	tags varCHAR(40)[] NOT NULL DEFAULT '{}',
	PRIMARY KEY (id)
) WITHOUT OIDS;
CREATE INDEX idx_meta_created ON meta_template (status, created) ;
CREATE INDEX idx_meta_size ON meta_template (size) ;
CREATE INDEX idx_meta_tags ON meta_template (tags) ;

CREATE TABLE meta_common
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

CREATE TABLE meta__deleted
(
	LIKE meta_template INCLUDING ALL,
	deleted timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) WITHOUT OIDS;

-- entry presave
CREATE TABLE meta__prepared (
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;


CREATE TABLE tag(
	id serial ,
	tag varchar(80) NOT NULL UNIQUE,
	item_count int NOT NULL DEFAULT 0,
	PRIMARY KEY  (id)
);


END;
