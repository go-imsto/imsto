
BEGIN;

set search_path = imsto, public;

CREATE TABLE meta_demo
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

CREATE TABLE meta_s3
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

CREATE TABLE meta_grid
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

END;
