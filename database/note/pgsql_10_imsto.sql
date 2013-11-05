-- 查询有扩展名是 .png 的 entry
select id, name from meta_wpitem where meta @> 'ext=>.png';
