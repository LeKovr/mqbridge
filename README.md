# mqbridge

> Stream messages from PG/NATS/File channel to another PG/NATS/File channel

[![Go Reference][ref1]][ref2]
 [![GitHub Release][gr1]][gr2]
 [![Build Status][bs1]][bs2]
 [![FOSSA Status][fs1]][fs2]
 [![GitHub license][gl1]][gl2]

[![codecov][cc1]][cc2]
 [![Test Coverage][cct1]][cct2]
 [![Maintainability][ccm1]][ccm2]
 [![GoCard][gc1]][gc2]

[cct1]: https://api.codeclimate.com/v1/badges/0f7f2d4314b28b45e333/test_coverage
[cct2]: https://codeclimate.com/github/LeKovr/mqbridge/test_coverage
[ccm1]: https://api.codeclimate.com/v1/badges/0f7f2d4314b28b45e333/maintainability
[ccm2]: https://codeclimate.com/github/LeKovr/mqbridge/maintainability
[fs1]: https://app.fossa.com/api/projects/git%2Bgithub.com%2FLeKovr%2Fmqbridge.svg?type=shield
[fs2]: https://app.fossa.com/projects/git%2Bgithub.com%2FLeKovr%2Fmqbridge?ref=badge_shield
[ref1]: https://pkg.go.dev/badge/github.com/LeKovr/mqbridge.svg
[ref2]: https://pkg.go.dev/github.com/LeKovr/mqbridge
[cc1]: https://codecov.io/gh/LeKovr/mqbridge/branch/master/graph/badge.svg
[cc2]: https://codecov.io/gh/LeKovr/mqbridge
[gc1]: https://goreportcard.com/badge/github.com/LeKovr/mqbridge
[gc2]: https://goreportcard.com/report/github.com/LeKovr/mqbridge
[bs1]: https://cloud.drone.io/api/badges/LeKovr/mqbridge/status.svg
[bs2]: https://cloud.drone.io/LeKovr/mqbridge
[gr1]: https://img.shields.io/github/release/LeKovr/mqbridge.svg
[gr2]: https://github.com/LeKovr/mqbridge/releases
[gl1]: https://img.shields.io/github/license/LeKovr/mqbridge.svg
[gl2]: https://github.com/LeKovr/mqbridge/blob/master/LICENSE

![Data flow](mqbridge.png)

Sending string from source to destination is a very primitive task. But it's solution might include some usefull technics. MQBridge uses this case for attempting to be example of doing things right.

## Message queue systems supported

  type | producer | consumer
-------|----------|----------
 file  | tail(file) | println file, data
  pg   | listen 'event' | select func(data)
  nats | Subscribe(channel) | Publish(channel, data)

## Installation

* Linux: just download & run. See [Latest release](https://github.com/LeKovr/mqbridge/releases/latest)
* Docker: `docker pull ghcr.io/lekovr/mqbridge`

## Plugins

**mqbridge** has 2 build modes - standalone (default) and **plugin** mode.

To activate **plugin** mode, run

```sh
make plugin-on
make run
```

You will see `Loading plugin` in program output (using `make run`).

To deactivate **plugin** mode, run

```sh
make clean plugin-off
```

## Config

```sh
  --point=       Endpoints connect string in form 'tag[:plugin[://dsn]]' (default: io:file)
  --bridge=      Bridge in form 'in_tag:in_channel:out_tag[:out_channel]' (default: io:src.txt,io:dst.txt)
```

### Connect strings

* **file** - `${TAG):file://`
* **pg** - `${TAG}:pg:postgres://user:pass@host:port/db?sslmode=disable`
* **nats** - `${TAG}:nats:nats://user:pass@host:port`

Where `${TAG}` is the name of endpoint given in config.

## Usage

### Producers

mqbridge uses the following as data source:

* **file** - tail files named as `in_channel`
* **pg** - listen `in_channel`
* **nats** - subscribe to `in_channel`

### Consumers

mqbridge sends received messages as

* **file** - add lines to file named as `out_channel`
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

```sh
./mqbridge --bridge in:event,out:bridge \
  --point in:pg:postgres://op:op@localhost:5432/db0?sslmode=disable \
  --point out:pg:postgres://op:op@localhost:5432/db1?sslmode=disable
```

3. Run at pg producer db (db0)

```sql
notify event, '{"test": 1972}';
```

4. See results in consumer db (db1)

```
select * from mqbridge_data ;
      line     
----------------
 {"test": 1972}
```

See also: [Examples directory](examples/)

## License

The MIT License (MIT), see [LICENSE](LICENSE).

Copyright (c) 2017-2021 Aleksey Kovrizhkin <lekovr+mqbridge@gmail.com>
