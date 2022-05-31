package config

import "strconv"

// DBConfig configures a database connection.
type DBConfig struct {
	Host    string
	Port    int
	User    string
	Pass    string
	Name    string
	Options string
}

// Config is the configuration for the server.
type Config struct {
	User  string
	Pass  string
	Port  int
	MySQL DBConfig
	TiDB  DBConfig
}

// DSN returns the data source name for the given database.
func (c *DBConfig) DSN() string {
	return c.User + ":" + c.Pass + "@tcp(" + c.Host + ":" + strconv.Itoa(c.Port) + ")/" + c.Name + "?" + c.Options
}

// Address returns the address for the given database.
func (c *DBConfig) Address() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}
