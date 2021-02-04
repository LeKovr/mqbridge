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

![Data flow](draw.io.dia.png)

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

## Application config

```sh
$ ./mqbridge -h
Usage:
  mqbridge [OPTIONS]

Application Options:
  --version      Show version and exit
  --debug        Show debug data

MQ Bridge Options:
  --delim=       Bridge definition delimiter (default: ,)
  --path_format= Path format (default: ./%s.so)
  --point=       Endpoints connect string in form 'tag[:plugin[://dsn]]' (default: io:file)
  --bridge=      Bridge in form 'in_tag:in_channel:out_tag[:out_channel]' (default: io:src.txt,io:dst.txt)

Help Options:
  -h, --help         Show this help message
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
create table mqbridge_data (line text);

create or replace function bridge(a text) returns void language plpgsql as
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

```sql
select * from mqbridge_data ;
      line     
----------------
 {"test": 1972}
```

See also: [Examples directory](examples/)

## Developer notes

This project solves quite simple task, but it shows some more important things named tests.
There are the following tests presented here:

* [stupid testing for unmockable package](plugins/pg/pg_test.go#L31)
* [tests with gomock](plugins/nats/nats_test.go)
* [golang test which runs docker container](plugins/pg/docker_test.go)
* [tests which use prepared environment](Makefile#L86)

### Makefile

Project contains `Makefile` with the following targets:

```bash
$ make

mqbridge Makefile: Stream messages from PG/NATS/File channel to another PG/NATS/File channel

Compile operations
    lint            run `golint` and `golangci-lint`
    vet             run `go vet`
    test            run tests
    test-docker     run tests that use services from docker-compose.yml
    test-docker-self run tests that run docker themselves
    cov-html        show package coverage in html (make cov-html PKG=counter)
    build           build app
    run             build and run in foreground

Plugin support
    plugin-on       enable plugin mode (this command changes source files)
    plugin-off      disable plugin mode (this command changes source files back)

Prepare distros
    build-standalone build like docker image from scratch
    buildall        build app for all platforms
    dist            create disro files

Docker operations
    up              start service in container
    down            stop service
    docker-build    build docker image
    docker-clean    remove docker image & temp files
    dc              run $(CMD) via docker-compose

Other
    clean           clean generated files
    update-godoc    update docs at pkg.go.dev
    update-ghcr     update latest docker image tag at ghcr.io
    help            list Makefile targets (this is default target)

```
## Project stats

Here is the data from [cloc](https://github.com/AlDanial/cloc)

### Real code (it does the real job)

Language|files|blank|comment|code
:-------|-------:|-------:|-------:|-------:
Go |12|97|63|677

Made by
```
cloc --not-match-f=test.go --md . | grep -E "Lang|:--|Go"
```

### All sources

Language|files|blank|comment|code
:-------|-------:|-------:|-------:|-------:
Go|23|200|105|1167
YAML|4|22|12|237
make|2|84|99|205
Markdown|2|68|0|200
Dockerfile|1|7|4|13
SQL|1|1|0|5
XML|1|0|0|1
--------|--------|--------|--------|--------
SUM:|34|382|220|1828


Made by
```
cloc --md .
```

## License

The MIT License (MIT), see [LICENSE](LICENSE).

Copyright (c) 2017-2021 Aleksey Kovrizhkin <lekovr+mqbridge@gmail.com>
