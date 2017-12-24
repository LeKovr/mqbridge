# mqbridge
> Translate messages from one message queue system to another one

[![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 [![GitHub code size in bytes][sz]]()
 [![GitHub license][gl1]][gl2]


[gc1]: https://goreportcard.com/badge/LeKovr/mqbridge
[gc2]: https://goreportcard.com/report/github.com/LeKovr/mqbridge
[gr1]: https://img.shields.io/github/release/LeKovr/mqbridge.svg
[gr2]: https://github.com/LeKovr/mqbridge/releases
[sz]: https://img.shields.io/github/languages/code-size/LeKovr/mqbridge.svg
[gl1]: https://img.shields.io/github/license/LeKovr/mqbridge.svg
[gl2]: LICENSE

![Data flow](mqbridge.png)

## Message queue systems supported

  type | producer | consumer
-------|----------|----------
 file  | tail(in_channel) | println out_channel, data
  pg   | listen in_channel | select out_channel(data)
  nats | Subscribe(in_channel) | Publish(out_channel, data)

## Installation

* Linux: just download & run. See [Latest release](https://github.com/LeKovr/mqbridge/releases/latest)
* Docker: `docker pull lekovr/mqbridge`. See [Docker store](https://store.docker.com/community/images/lekovr/mqbridge)

## Config

```
      --in=        Producer connect string
      --out=       Consumer connect string
      --bridge=    Bridge(s) in form 'in_channel[,out_channel]'
```

### Connect strings

* **file** - `file://`
* **pg** - `postgres://user:pass@host:port/db?sslmode=disable`
* **nats** - `nats://user:pass@host:port`

## Usage

### Producers

See also: [Examples directory](examples/)

mqbridge uses the following as data source:

* **file** - tail files named as `in_channel`
* **pg** - listen `in_channel`
* **nats** - subscribe to `in_channel`

### Consumers

mqbridge sends received messages as

* **file** - add lines to files named as `out_channel`
* **pg** - calls sql `select out_channel(data)`
* **nats** - publish message to `out_channel`

### pg usage sample

This sample shows how to setup pg -> pg bridge.

1. Setup pg consumer (db1) for `out_channel` = `bridge` (see function name)
```sql
create table mqbridge_data (line jsonb);

create or replace function bridge(a jsonb) returns void language plpgsql as 
$_$ 
  begin insert into mqbridge_data (line) values(a); end 
$_$;
```
2. Run mqbridge
```
./mqbridge --bridge event,bridge \
  --in postgres://op:op@localhost:5432/db0?sslmode=disable \
 --out postgres://op:op@localhost:5432/db1?sslmode=disable
```
3. Run at pg producer (db0) SQL
```sql
notify event, '{"test": 1972}';
```
4. See results in db1
```
select * from mqbridge_data ;
      line     
----------------
 {"test": 1972}

```

## TODO

* [ ] centrifugo support
* [ ] add channel buffer length in bridge config args (--buffer []int)
* [ ] (may be) postgresql: reuse connect if in=out
* [ ] (may be) use github.com/jaehue/anyq for mq bridge

## License

The MIT License (MIT), see [LICENSE](LICENSE).

Copyright (c) 2017 Alexey Kovrizhkin <lekovr+mqbridge@gmail.com>
