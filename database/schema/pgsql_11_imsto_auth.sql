
-- 访问授权相关
BEGIN;

set search_path = imsto, public;

CREATE TABLE upload_ticket (
	id serial,
	app_id int NOT NULL DEFAULT 0,
	author int NOT NULL ,
	prompt varchar(255) NOT NULL DEFAULT '',
	url_prefix varchar(112) NOT NULL DEFAULT '',
	img_id varchar (44) NOT NULL DEFAULT '',
	img_path varchar(65) NOT NULL DEFAULT '',
	img_meta hstore,
	uploaded boolean NOT NULL DEFAULT false,
	created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) WITHOUT OIDS;


END;
