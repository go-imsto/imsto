-- 查询有扩展名是 .png 的 entry
SELECT id, name FROM meta_wpitem WHERE meta @> 'ext=>.png';

CREATE TABLE meta_s3
(
	LIKE meta_template INCLUDING ALL
) WITHOUT OIDS;

CREATE OR REPLACE VIEW v_crc AS
SELECT count(id) as nn, crc64, max(size) as s1, min(size) as s2, max(path) p1, min(path) p2
FROM meta_crc
GROUP BY crc64;

SELECT * FROM v_crc WHERE s1 != s2;
SELECT count(*) FROM v_crc WHERE nn > 1;
SELECT * FROM v_crc WHERE nn > 1;
ORDER BY nn DESC LIMIT 20;

 nn |      crc64       | max_size | min_size
----+------------------+----------+----------
  2 | 0049262201ced289 |    52828 |    52828
  2 | 00b352105cd0cc49 |    43923 |    43923
  2 | 00297eb4fe7f4993 |    96260 |    96260
  2 | 00e54844bc63b8c2 |    21239 |    21239
  2 | 00d958ef857e24e5 |    70693 |    70693

