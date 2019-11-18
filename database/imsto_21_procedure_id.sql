
BEGIN;

set search_path = imsto, public;

CREATE SCHEMA IF NOT EXISTS shard_1;
CREATE SEQUENCE IF NOT EXISTS shard_1.global_id_sequence;

CREATE OR REPLACE FUNCTION shard_1.id_generator(OUT result bigint) AS $$
DECLARE
    our_epoch CONSTANT bigint := 1502725500000;  -- '2017-08-14 15:45:00'
    seq_id bigint;
    now_millis bigint;
    -- the id of this DB shard, must be set for each
    -- schema shard you have - you could pass this as a parameter too
    shard_id int := 1;
BEGIN
    SELECT nextval('shard_1.global_id_sequence') % 2048 INTO seq_id;

    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    result := (now_millis - our_epoch) << 21;
    result := result | (shard_id << 11);
    result := result | (seq_id);
END;
$$ LANGUAGE PLPGSQL;

-- SELECT shard_1.id_generator()

/*
ms := int64(id>>21) + epoch
	sec := ms / 1000
	tm = time.Unix(sec, (ms-sec*1000)*int64(time.Millisecond))
	shardId = (id >> 11) & shardMask
	seqId = id & seqMask
*/
CREATE OR REPLACE FUNCTION id_split(id bigint, OUT tm timestamp, OUT shard_id int, OUT seq_id int) AS $$
DECLARE
    our_epoch CONSTANT bigint := 1502725500000;  -- '2017-08-14 15:45:00'
    ms bigint;
BEGIN
	ms := (id >> 21) + our_epoch;
	tm := to_timestamp(ms / 1000);
	shard_id := (id >> 11) & 1023;
	seq_id := id & 2047;
END;
$$ LANGUAGE PLPGSQL;


END;
