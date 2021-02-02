# mqbridge usage examples

All examples require [docker](http://docker.io) and `make` installed.

There are the following examples available:

* `test-file`: test file -> file bridge
* `test-nats`: test file -> nats -> file bridges
* `test-pg`: test file -> pg -> file bridges

## Sample output

```bash
$ make test-pg
** Clear src.txt and dst.txt **
Creating mqbridge_pg_1 ... done
Creating mqbridge_mqbr-pg_1 ... done
** Fill src.txt **
** Cat src.txt **
1000
1001
1002
1003
1004
1005
** Cat dst.txt **
1000
1001
1002
1003
1004
1005
** Shutdown services **
Stopping mqbridge_mqbr-pg_1 ... done
Stopping mqbridge_pg_1      ... done
Removing mqbridge_mqbr-pg_1 ... done
Removing mqbridge_pg_1      ... done
Going to remove mqbridge_mqbr-pg_1, mqbridge_pg_1
```
