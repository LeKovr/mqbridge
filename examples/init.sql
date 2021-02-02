\c mqbridge_example

create or replace function test_channel(a jsonb) returns void language plpgsql as
$_$
  BEGIN PERFORM pg_notify('test_event', a::text); END
$_$;
