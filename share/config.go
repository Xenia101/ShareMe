package share

import (
	"encoding/json"
	"io"
	"os"
)

var (
	Config struct {
		Listen      string `json:"listen"`
		MysqlSource string `json:"mysql_source"`
		SentryDsn   string `json:"sentry_dsn"`
	}
)

func init() {
	fs, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	err = json.NewDecoder(fs).Decode(&Config)
	if err != nil && err != io.EOF {
		panic(err)
	}
}
