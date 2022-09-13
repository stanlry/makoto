package main

type dbConfig struct {
	Postgres postgres `toml:"postgres"`
}

type postgres struct {
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	DBName   string `toml:"name"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	SSLMode  string `toml:"ssl_mode"`
}
