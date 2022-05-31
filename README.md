# TiCompare

Compare query results between MySQL and TiDB server.

## Usage

```
➜  ticomp git:(master) ✗ ./ticomp -h            
Usage:
  ticomp [flags]

Flags:
  -h, --help                   help for ticomp
      --mysql.host string      MySQL server host name (default "127.0.0.1")
      --mysql.name string      MySQL server database name
      --mysql.options string   MySQL server connection options (default "charset=utf8mb4")
      --mysql.pass string      MySQL server password
      --mysql.port int         MySQL server port (default 3306)
      --mysql.user string      MySQL server user name (default "root")
      --pass string            TiCompare shadow server password
  -P, --port int               Listen port of TiCompare shadow server (default 5001)
      --tidb.host string       TiDB server host name (default "127.0.0.1")
      --tidb.name string       TiDB server database name
      --tidb.options string    TiDB server connection options (default "charset=utf8mb4")
      --tidb.pass string       TiDB server password
      --tidb.port int          TiDB server port (default 4000)
      --tidb.user string       TiDB server user name (default "root")
      --user string            TiCompare shadow server user name (default "root")

```

###

1. Run TiComp and connect to local MySQL/TiDB server

    ```shell
    # Login local mysql server with user name: lonng
    ./ticomp --port 6000 --mysql.user lonng
    ```
   
2. Connect to TiComp and treat it as a normal MySQL server

    ```shell
    # Login into TiCompare server
    mysql -h 127.0.0.1 -P 6000 -uroot

    # Query
    mysql> select uuid();
    +--------------------------------------+
    | uuid()                               |
    +--------------------------------------+
    | bbfb289e-e125-11ec-b832-c8f6766ec590 |
    +--------------------------------------+
    1 row in set (0.01 sec)
    ```

3. Check your TiCompare output and it should be like following content with diff highlight

    ```
    QUERY >	 select uuid()
    TiDB  >
    +--------------------------------------+
    | uuid()                               |
    +--------------------------------------+
    | bbfb6642-e125-11ec-846a-acde48001122 |
    +--------------------------------------+
    MySQL >
    +--------------------------------------+
    | uuid()                               |
    +--------------------------------------+
    | bbfb289e-e125-11ec-b832-c8f6766ec590 |
    +--------------------------------------+
    ```
