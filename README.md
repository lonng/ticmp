# TiCmp

Compare query results between MySQL and TiDB server.

## Usage

```
➜  ticmp git:(master) ✗ ./ticmp -h            
Usage:
  ticmp [flags]

Flags:
  -P, --port int               Listen port of TiCmp shadow server (default 5001)
      --user string            TiCmp shadow server user name (default "root")
      --pass string            TiCmp shadow server password
      --html string            Output compare to specified html file
      --csv string             Output compare to specified csv file
      --mysql.host string      MySQL server host name (default "127.0.0.1")
      --mysql.port int         MySQL server port (default 3306)
      --mysql.user string      MySQL server user name (default "root")
      --mysql.pass string      MySQL server password
      --mysql.name string      MySQL server database name
      --mysql.options string   MySQL server connection options (default "charset=utf8mb4")
      --tidb.host string       TiDB server host name (default "127.0.0.1")
      --tidb.port int          TiDB server port (default 4000)
      --tidb.user string       TiDB server user name (default "root")
      --tidb.pass string       TiDB server password
      --tidb.name string       TiDB server database name
      --tidb.options string    TiDB server connection options (default "charset=utf8mb4")
  -h, --help                   help for ticmp
```

###

1. Run ticmp and connect to local MySQL/TiDB server

    ```shell
    # Login local mysql server with user name: lonng
    ./ticmp --port 6000 --mysql.user lonng
    ```
   
2. Connect to ticmp and treat it as a normal MySQL server

    ```shell
    # Login into TiCmp server
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

3. Check your TiCmp output and it should be like following content with diff highlight

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
