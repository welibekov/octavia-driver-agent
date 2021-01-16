package config

import (
	"fmt"
	"strings"
	"gopkg.in/ini.v1"
)

type Url struct {
	Database string
	Rabbit string
}

func Get(configFile string) (error, Url) {
	var url Url

	cfg, err := ini.Load(configFile)
	if err != nil {
		return err,url
	}
	amqp_string := cfg.Section("").Key("transport_url").String()
	db_string := cfg.Section("database").Key("connection").String()

	url.Rabbit = strings.Replace(amqp_string, "rabbit://", "amqp://", 1)

	db_slice := strings.Split(db_string, ":")
	db_user := strings.TrimLeft(db_slice[1], "//")
	db_name := strings.Split(db_slice[len(db_slice)-1], "/")[1]
	db_host := strings.Split(
				strings.Split(
				db_slice[len(db_slice)-1], "@")[1],
				"/")[0]
	db_pass := strings.Split(db_slice[len(db_slice)-1],"@")[0]
	url.Database = fmt.Sprintf("%s:%s@tcp(%s)/%s",db_user,db_pass,db_host,db_name)
	return nil,url
}
