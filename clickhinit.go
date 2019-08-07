package clickhouse

import (
	"database/sql"
	_ "github.com/kshvakov/clickhouse"
	"github.com/obase/conf"
	"net/url"
	"strconv"
	"strings"
)

const CKEY = "clickhouse"

// 对接conf.yml, 读取原mysql相关配置
func init() {
	configs, ok := conf.GetSlice(CKEY)
	if !ok || len(configs) == 0 {
		return
	}

	for _, config := range configs {
		if key, ok := conf.ElemString(config, "key"); ok {
			sb := new(strings.Builder)

			username, ok := conf.ElemString(config, "username")
			if ok {
				addit(sb, "username", username)
			}
			password, ok := conf.ElemString(config, "password")
			if ok {
				addit(sb, "password", password)
			}
			database, ok := conf.ElemString(config, "database")
			if ok {
				addit(sb, "database", database)
			}
			read_timeout, ok := conf.ElemString(config, "read_timeout")
			if ok {
				addit(sb, "read_timeout", read_timeout)
			}
			write_timeout, ok := conf.ElemString(config, "write_timeout")
			if ok {
				addit(sb, "write_timeout", write_timeout)
			}
			no_delay, ok := conf.ElemString(config, "no_delay")
			if ok {
				addit(sb, "no_delay", no_delay)
			}
			connection_open_strategy, ok := conf.ElemString(config, "connection_open_strategy")
			if ok {
				addit(sb, "connection_open_strategy", connection_open_strategy)
			}
			block_size, ok := conf.ElemString(config, "block_size")
			if ok {
				addit(sb, "block_size", block_size)
			}
			pool_size, ok := conf.ElemString(config, "pool_size")
			if ok {
				addit(sb, "pool_size", pool_size)
			}
			debug, ok := conf.ElemString(config, "debug")
			if ok {
				addit(sb, "debug", debug)
			}
			address, ok := conf.ElemStringSlice(config, "address")
			if len(address) > 1 {
				addit(sb, "alt_hosts", strings.Join(address[1:], ","))
			}

			maxIdleConns, ok := conf.ElemInt(config, "maxIdleConns")
			if !ok {
				maxIdleConns, _ = strconv.Atoi(pool_size)
			}
			maxOpenConns, ok := conf.ElemInt(config, "maxOpenConns")
			if !ok {
				maxOpenConns, _ = strconv.Atoi(pool_size)
			}
			connMaxLifetime, ok := conf.ElemDuration(config, "connMaxLifetime")

			defalt, ok := conf.ElemBool(config, "default")

			db, err := sql.Open("clickhouse", "tcp://"+address[0]+sb.String())
			if err != nil {
				panic(err)
			}
			if maxIdleConns > 0 {
				db.SetMaxIdleConns(maxIdleConns)
			}
			if maxOpenConns > 0 {
				db.SetMaxOpenConns(maxOpenConns)
			}
			if connMaxLifetime > 0 {
				db.SetConnMaxLifetime(connMaxLifetime)
			}

			err = Setup(key, db, defalt)
			if err != nil {
				panic(err)
			}
		}
	}
}

func addit(sb *strings.Builder, key, val string) {

	if strings.TrimSpace(val) == "" {
		return
	}

	if sb.Len() == 0 {
		sb.WriteByte('?')
	} else {
		sb.WriteByte('&')
	}
	sb.WriteString(key)
	sb.WriteByte('=')
	sb.WriteString(url.QueryEscape(val))
}
