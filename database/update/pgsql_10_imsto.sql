ALTER TABLE map_template ADD roofs varCHAR(12)[] NOT NULL DEFAULT '{}';
ALTER TABLE map_template ADD created timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE upload_ticket RENAME section to roof;


-- 20131216
BEGIN;
UPDATE map_template SET sev = '' WHERE sev IS NULL;
ALTER TABLE map_template ALTER sev SET DEFAULT '';
ALTER TABLE map_template ALTER sev SET NOT NULL;

UPDATE meta_template SET sev = '' WHERE sev IS NULL;
ALTER TABLE meta_template ALTER sev SET DEFAULT '';
ALTER TABLE meta_template ALTER sev SET NOT NULL;

UPDATE meta_template SET meta = '' WHERE meta IS NULL;
ALTER TABLE meta_template ALTER meta SET DEFAULT '';
ALTER TABLE meta_template ALTER meta SET NOT NULL;

UPDATE meta_template SET exif = '' WHERE exif IS NULL;
ALTER TABLE meta_template ALTER exif SET DEFAULT '';
ALTER TABLE meta_template ALTER exif SET NOT NULL;


UPDATE prepared_entry SET meta = '' WHERE meta IS NULL;
ALTER TABLE prepared_entry ALTER meta SET DEFAULT '';
ALTER TABLE prepared_entry ALTER meta SET NOT NULL;

UPDATE prepared_entry SET exif = '' WHERE exif IS NULL;
ALTER TABLE prepared_entry ALTER exif SET DEFAULT '';
ALTER TABLE prepared_entry ALTER exif SET NOT NULL;

END;

-- 20131225
ALTER TABLE meta_template ADD roof varCHAR(12) NOT NULL DEFAULT '';
CREATE INDEX idx_meta_meta ON meta_template (meta) ;
CREATE INDEX idx_meta_size ON meta_template (size) ;

BEGIN;
ALTER TABLE meta_demo ADD roof varCHAR(12) NOT NULL DEFAULT '';
UPDATE meta_demo SET meta = meta || hstore('name', name) WHERE not meta ? 'name';
CREATE INDEX ON meta_demo (meta) ;
CREATE INDEX ON meta_demo (size) ;
END;

UPDATE meta__prepared SET size = (meta -> 'size')::int WHERE size = 0;

-- for i in {1..30}; do echo $i && imsto import -ready; done;

ALTER TABLE meta_template ADD tags varCHAR(40)[] NOT NULL DEFAULT '{}';
CREATE INDEX idx_meta_tags ON meta_template (tags) ;

ALTER TABLE meta__prepared ADD tags varCHAR(40)[] NOT NULL DEFAULT '{}';
ALTER TABLE meta__deleted ADD tags varCHAR(40)[] NOT NULL DEFAULT '{}';

ALTER TABLE meta_demo ADD tags varCHAR(40)[] NOT NULL DEFAULT '{}';
CREATE INDEX ON meta_demo (tags, status) ;

BEGIN;
ALTER DOMAIN entry_path
  DROP CONSTRAINT entry_path_check;
ALTER DOMAIN entry_path
  ADD CHECK (VALUE ~ '^[a-z0-9]{2}/?[a-z0-9]{2}/?[a-z0-9]{5,32}\.[a-z0-9]{2,6}$');
END;

BEGIN;
ALTER DOMAIN entry_id
  DROP CONSTRAINT entry_id_check;
ALTER DOMAIN entry_id
  ADD CHECK (VALUE ~ '^[a-z0-9]{9,36}$');
END;

ALTER FUNCTION tag_add(text, text, text[]) RENAME TO tag_map;
ALTER FUNCTION tag_remove(text, text, text[]) RENAME TO tag_unmap;

ALTER DOMAIN entry_id RENAME TO entry_xid;

ALTER TABLE meta_template ADD UNIQUE (hashes);
ALTER TABLE meta__prepared ADD UNIQUE (hashes);
ALTER TABLE meta_demo ADD UNIQUE (hashes);
