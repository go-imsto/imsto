-- pgsql_10_imsto.sql

-- require: >= postgresql-9.4;

BEGIN;

CREATE SCHEMA imsto;
COMMENT ON SCHEMA imsto IS '存储相关';
GRANT ALL ON SCHEMA imsto TO imsto;

set search_path = imsto, public;

CREATE DOMAIN entry_xid AS TEXT
CHECK(
	VALUE ~ '^[a-z0-9]{9,36}$'
);

CREATE DOMAIN entry_path AS TEXT
CHECK (
	VALUE ~ '^[a-z0-9]{2}/?[a-z0-9]{2}/?[a-z0-9]{5,32}\.[a-z0-9]{2,6}$'
);


-- all file hash values
CREATE TABLE hash_template (
	hashed varCHAR(40) NOT NULL  ,
	item_id entry_xid NOT NULL ,
	-- prefix varCHAR(10) NOT NULL DEFAULT '' ,
	path entry_path NOT NULL ,
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (hashed)
) WITHOUT OIDS;

-- mapping for id and storage engine item
CREATE TABLE map_template (
	id entry_xid NOT NULL,
	name varCHAR(120) NOT NULL DEFAULT '',
	path entry_path NOT NULL ,
	size int NOT NULL DEFAULT 0 CHECK (size >= 0),
	sev jsonb NOT NULL DEFAULT '{}'::jsonb, -- storage info
	status smallint NOT NULL DEFAULT 0, -- 0=valid,1=deleted
	roofs varCHAR(12)[] NOT NULL DEFAULT '{}',
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) WITHOUT OIDS;

-- meta browsable
CREATE TABLE meta_template (
	id entry_xid NOT NULL,
	path entry_path NOT NULL ,
	name varCHAR(120) NOT NULL DEFAULT '',
	roof varCHAR(12) NOT NULL DEFAULT '',
	meta jsonb NOT NULL DEFAULT '{}'::jsonb,
	hashes varCHAR(40)[],
	ids varCHAR(38)[],
	size int NOT NULL DEFAULT 0,
	sev jsonb NOT NULL DEFAULT '{}'::jsonb,
	exif jsonb NOT NULL DEFAULT '{}'::jsonb, -- exif info
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
	label varchar(80) NOT NULL UNIQUE,
	item_count int NOT NULL DEFAULT 0,
	PRIMARY KEY  (id)
);


END;
