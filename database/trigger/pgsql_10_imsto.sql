-- pgsql_061_content_storage.sql

BEGIN;

set search_path = imsto;


CREATE TRIGGER hash_insert_trigger BEFORE INSERT ON hash_template
FOR EACH ROW EXECUTE PROCEDURE hash_insert_trigger();


END;
