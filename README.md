# mqbridge
Translate messages from one message queue system to another one

## Message queue systems supported

  type | in | out
-------|----|-----
 file  | tail(channelIn) | print channelOut,data
  pg   | listen channelIn | select channelOut(data)
  nats | Subscribe(channelIn) | Publish(channelOut,data)

## Config

```
      --in=        Side 'in' connect string
      --out=       Side 'out' connect string
      --bridge=    Bridge(s) in form 'in_channel[:out_channel]'
```

## Usage

```sql
create table mqbridge_data (line jsonb);
create or replace function bridge(a jsonb) returns void language plpgsql as $_$ begin insert into mqbridge_data (line) values(a); end $_$;
```

```sh
C=postgres://op:op@localhost:5432/op?sslmode=disable
F=file://
D=nats://localhost:4222
go run *.go --bridge event:bridge --in $C --out $F --log_level debug
```

Now just run
```sql
notify event, '{"test": 4}';
```

## TODO

* [ ] centrifugo support
* [ ] (may be) postgresql: reuse connect if in=out
* [ ] (may be) use github.com/jaehue/anyq for mq bridge

