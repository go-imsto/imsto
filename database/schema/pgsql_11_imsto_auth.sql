
-- 访问授权相关
BEGIN;

set search_path = imsto, public;

CREATE TABLE apps (
	id serial,
	api_key varCHAR(120) NOT NULL UNIQUE,
	name name NOT NULL,
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) WITHOUT OIDS;

CREATE TABLE upload_ticket (
	id serial,
	roof varchar(20) NOT NULL,
	app_id smallint NOT NULL,
	author int NOT NULL ,
	prompt varchar(255) NOT NULL,
	url_prefix varchar(112) NOT NULL DEFAULT '',
	img_id varchar (44) NOT NULL DEFAULT '',
	img_path varchar(65) NOT NULL DEFAULT '',
	img_meta hstore,
	done boolean NOT NULL DEFAULT false,
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) WITHOUT OIDS;

CREATE INDEX idx_ticket_created ON upload_ticket (created, done) ;

SELECT setval('upload_ticket_id_seq', 1000, true);


END;
