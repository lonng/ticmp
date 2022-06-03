package main

import (
	"fmt"
	"net"
	"os"

	"github.com/go-mysql-org/go-mysql/server"
	"github.com/lonng/ticomp/config"
	"github.com/lonng/ticomp/handler"
	"github.com/lonng/ticomp/render"
	"github.com/spf13/cobra"
)

func onConnect(c net.Conn, cfg *config.Config, rndr render.Render) error {
	connIdent := c.RemoteAddr().String()
	h := handler.NewShadowHandler(cfg, connIdent, rndr)

	err := h.Initialize()
	if err != nil {
		return err
	}

	defer func() {
		if err := h.Finalize(); err != nil {
			fmt.Println("Finalize shadow handler failed", err)
		}
	}()

	// Create a connection with user root and an empty password.
	// You can use your own handler to handle command here.
	conn, err := server.NewConn(c, cfg.User, cfg.Pass, h)
	if err != nil {
		return err
	}

	// as long as the client keeps sending commands, keep handling them
	for {
		if err := conn.HandleCommand(); err != nil {
			fmt.Printf("handle command error: %v\n", err)
			return err
		}
	}
}

func main() {
	cfg := &config.Config{}
	cmd := &cobra.Command{
		Use:          "ticomp",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var rndr render.Render
			if cfg.HTMLPath != "" {
				hr := &render.HTMLRender{}
				if err := hr.Open(cfg.HTMLPath); err != nil {
					return err
				}
				rndr = hr
				defer hr.Close()
			} else if cfg.CSVPath != "" {
				hr := &render.CSVRender{}
				if err := hr.Open(cfg.CSVPath); err != nil {
					return err
				}
				rndr = hr
				defer hr.Close()
			} else {
				rndr = render.ConsoleRender{}
			}

			// Wrap with asynchronous render
			rd := render.NewAsyncRender(rndr)
			go rd.Start()

			l, err := net.Listen("tcp4", fmt.Sprintf(":%d", cfg.Port))
			if err != nil {
				return err
			}

			fmt.Printf("Serve successfully (mysql -h 127.0.0.1 -P %d -uroot)\n", cfg.Port)

			// Accept a new connection once
			for {
				c, err := l.Accept()
				if err != nil {
					return err
				}

				go onConnect(c, cfg, rd)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	// Shadow server configurations
	flags.IntVarP(&cfg.Port, "port", "P", 5001, "Listen port of TiCompare shadow server")
	flags.StringVar(&cfg.User, "user", "root", "TiCompare shadow server user name")
	flags.StringVar(&cfg.Pass, "pass", "", "TiCompare shadow server password")
	flags.StringVar(&cfg.HTMLPath, "html", "", "Output compare to specified html file")
	flags.StringVar(&cfg.CSVPath, "csv", "", "Output compare to specified html file")

	// MySQL server configurations
	flags.StringVar(&cfg.MySQL.Host, "mysql.host", "127.0.0.1", "MySQL server host name")
	flags.IntVar(&cfg.MySQL.Port, "mysql.port", 3306, "MySQL server port")
	flags.StringVar(&cfg.MySQL.User, "mysql.user", "root", "MySQL server user name")
	flags.StringVar(&cfg.MySQL.Pass, "mysql.pass", "", "MySQL server password")
	flags.StringVar(&cfg.MySQL.Name, "mysql.name", "", "MySQL server database name")
	flags.StringVar(&cfg.MySQL.Options, "mysql.options", "charset=utf8mb4", "MySQL server connection options")

	// TiDB server configurations
	flags.StringVar(&cfg.TiDB.Host, "tidb.host", "127.0.0.1", "TiDB server host name")
	flags.IntVar(&cfg.TiDB.Port, "tidb.port", 4000, "TiDB server port")
	flags.StringVar(&cfg.TiDB.User, "tidb.user", "root", "TiDB server user name")
	flags.StringVar(&cfg.TiDB.Pass, "tidb.pass", "", "TiDB server password")
	flags.StringVar(&cfg.TiDB.Name, "tidb.name", "", "TiDB server database name")
	flags.StringVar(&cfg.TiDB.Options, "tidb.options", "charset=utf8mb4", "TiDB server connection options")

	if err := cmd.Execute(); err != nil {
		fmt.Println("Execute command failed", err)
		os.Exit(2)
	}
}
