
\encoding UTF8

BEGIN;

-- （参观者）只读角色
CREATE ROLE im_visitor CONNECTION LIMIT 8;
ALTER ROLE im_visitor SET client_encoding=utf8;
-- （维护者）读写角色
CREATE ROLE im_keeper CONNECTION LIMIT 6;
ALTER ROLE im_keeper SET client_encoding=utf8;

-- 只读登录账户
CREATE ROLE im_reader LOGIN PASSWORD 'read0fromlink';
GRANT im_visitor TO im_reader;


-- 账户相关

CREATE ROLE im_storage LOGIN PASSWORD 'aside26,dicx';
CREATE DATABASE im_storage WITH OWNER = im_storage ENCODING = 'UTF8';
GRANT ALL ON DATABASE im_storage TO im_storage;
ALTER ROLE im_storage SET client_encoding=utf8;
GRANT CONNECT ON DATABASE im_storage TO public;
GRANT CONNECT, TEMPORARY ON DATABASE im_storage TO GROUP wp_keeper;
GRANT CONNECT, TEMPORARY ON DATABASE im_storage TO GROUP wp_visitor;


END;
