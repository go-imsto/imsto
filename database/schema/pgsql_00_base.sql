
\encoding UTF8

-- BEGIN;

-- visitor (readonly)
CREATE ROLE im_visitor CONNECTION LIMIT 8;
ALTER ROLE im_visitor SET client_encoding=utf8;
-- keeper (read, write)
CREATE ROLE im_keeper CONNECTION LIMIT 6;
ALTER ROLE im_keeper SET client_encoding=utf8;

-- readonly account
CREATE ROLE im_reader LOGIN PASSWORD 'read0fromlink';
GRANT im_visitor TO im_reader;


-- imsto database

CREATE ROLE imsto LOGIN PASSWORD 'aside26,dicx';
CREATE DATABASE imsto WITH OWNER = imsto ENCODING = 'UTF8';
GRANT ALL ON DATABASE imsto TO imsto;
ALTER ROLE imsto SET client_encoding=utf8;
GRANT CONNECT ON DATABASE imsto TO public;
GRANT CONNECT, TEMPORARY ON DATABASE imsto TO GROUP im_keeper;
GRANT CONNECT, TEMPORARY ON DATABASE imsto TO GROUP im_visitor;


--END;
