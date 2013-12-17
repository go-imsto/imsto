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

