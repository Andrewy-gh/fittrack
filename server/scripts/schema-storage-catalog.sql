-- Stable storage definitions for migration/schema.sql comparison; see ../schema-verification.md.
-- Names and SQL expressions come from PostgreSQL, not parsed migration text.

WITH objects AS (
SELECT 'table:'||c.relname AS key, jsonb_build_object('kind',c.relkind,'persistence',c.relpersistence) AS value
FROM pg_class c WHERE c.relnamespace='public'::regnamespace AND c.relkind IN ('r','p') AND c.relname<>'goose_db_version'
UNION ALL
SELECT 'column:'||c.relname||'.'||a.attname,
jsonb_build_object('type',format_type(a.atttypid,a.atttypmod),'not_null',a.attnotnull,'default',pg_get_expr(d.adbin,d.adrelid),'identity',a.attidentity,'generated',a.attgenerated,'collation',coll.collname)
FROM pg_class c JOIN pg_attribute a ON a.attrelid=c.oid LEFT JOIN pg_attrdef d ON d.adrelid=c.oid AND d.adnum=a.attnum LEFT JOIN pg_collation coll ON coll.oid=a.attcollation
WHERE c.relnamespace='public'::regnamespace AND c.relkind IN ('r','p') AND c.relname<>'goose_db_version' AND a.attnum>0 AND NOT a.attisdropped
UNION ALL
SELECT 'constraint:'||c.relname||'.'||con.conname,
jsonb_build_object('definition',pg_get_constraintdef(con.oid),'validated',con.convalidated,'deferrable',con.condeferrable,'deferred',con.condeferred)
FROM pg_constraint con JOIN pg_class c ON c.oid=con.conrelid WHERE c.relnamespace='public'::regnamespace AND c.relname<>'goose_db_version'
UNION ALL
SELECT 'index:'||c.relname||'.'||i.relname,
jsonb_build_object('definition',pg_get_indexdef(i.oid),'valid',x.indisvalid,'ready',x.indisready,'nulls_not_distinct',x.indnullsnotdistinct)
FROM pg_index x JOIN pg_class c ON c.oid=x.indrelid JOIN pg_class i ON i.oid=x.indexrelid WHERE c.relnamespace='public'::regnamespace AND c.relname<>'goose_db_version'
UNION ALL
SELECT 'sequence:'||c.relname,
jsonb_build_object('type',format_type(s.seqtypid,NULL),'start',s.seqstart,'increment',s.seqincrement,'min',s.seqmin,'max',s.seqmax,'cache',s.seqcache,'cycle',s.seqcycle,'owned_table',t.relname,'owned_column',a.attname)
FROM pg_sequence s JOIN pg_class c ON c.oid=s.seqrelid LEFT JOIN pg_depend d ON d.objid=c.oid AND d.deptype='a' LEFT JOIN pg_class t ON t.oid=d.refobjid LEFT JOIN pg_attribute a ON a.attrelid=t.oid AND a.attnum=d.refobjsubid
WHERE c.relnamespace='public'::regnamespace AND c.relname<>'goose_db_version_id_seq'
)
SELECT jsonb_object_agg(key,value) FROM objects;
