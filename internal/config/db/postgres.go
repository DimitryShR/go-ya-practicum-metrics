package db

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const defaultSSLMode = "disable"

type PgConn struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SslMode  string
}

func NewPgConn(host, port, user, password, dbName string, sslmode *string) (*PgConn, error) {
	intPort, err := parsePort(port)
	if err != nil {
		return nil, err
	}

	if sslmode == nil {
		defaultSslMode := defaultSSLMode
		sslmode = &defaultSslMode
	}

	return &PgConn{Host: host, Port: intPort, User: user, Password: password, DBName: dbName, SslMode: *sslmode}, nil
}

func NewPgConnDsn(dsn string) (*PgConn, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, nil
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}
	if u.Scheme == "" {
		return nil, fmt.Errorf("dsn must be URL format, got empty scheme")
	}

	var user, password string
	if u.User != nil {
		user = u.User.Username()
		password, _ = u.User.Password()
	}

	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("dsn must include host")
	}

	port, err := parsePort(u.Port())
	if err != nil {
		return nil, err
	}

	dbname := strings.TrimPrefix(u.Path, "/")

	sslmode := u.Query().Get("sslmode")
	if sslmode == "" {
		sslmode = defaultSSLMode
	}

	return &PgConn{Host: host, Port: port, User: user, Password: password, DBName: dbname, SslMode: sslmode}, nil
}

func parsePort(port string) (int, error) {
	if port == "" {
		return 5432, nil
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return 0, fmt.Errorf("invalid port: %w", err)
	}
	return p, nil
}

func (p PgConn) GetDsn() string {
	if p.Host == "" {
		return ""
	}
	if p.SslMode == "" {
		p.SslMode = defaultSSLMode
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SslMode)
}
