package share

import (
	"io"
	"os"
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

var (
	Config struct {
		HTTPListen  string `json:"http_listen"`
		MysqlSource string `json:"mysql_source"`
		SentryDsn   string `json:"sentry_dsn"`

		Dir struct {
			Public string `json:"public"`
			Upload string `json:"upload"`
		} `json:"dir"`

		Expires struct {
			Upload time.Duration `json:"upload"` // 업로드 제한시간
			Idle   time.Duration `json:"idle"`
			Forced time.Duration `json:"forced"`
		} `json:"expires"`

		Cleanup struct {
			Duration   time.Duration `json:"duration"`    // 갱신 주기
			QueryLimit int           `json:"query_limit"` // 1 select 문 당 가져올 limit 수
		} `json:"cleanup"`

		IDRule struct {
			Min      int             `json:"min"`      // 최소 길이
			Max      int             `json:"max"`      // 최대 길이
			Chars    string          `json:"chars"`    // 사용할 문자들
			Conflict []ConflictChars `json:"conflict"` // 겹치지 않게 할 문자들
		} `json:"id_rule"`
	}
)

type ConflictChars []byte

func init() {
	jsoniter.RegisterTypeDecoderFunc(
		"share.ConflictChars",
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			switch iter.WhatIsNext() {
			case jsoniter.StringValue:
				*(*ConflictChars)(ptr) = []byte(iter.ReadString())

			default:
				*(*interface{})(ptr) = iter.Read()
			}
		},
	)

	jsoniter.RegisterTypeDecoderFunc(
		"time.Duration",
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			switch iter.WhatIsNext() {
			case jsoniter.StringValue:
				td, err := time.ParseDuration(iter.ReadString())
				if err != nil {
					iter.Error = err
					return
				}

				*(*time.Duration)(ptr) = td

			default:
				*(*interface{})(ptr) = iter.Read()
			}
		},
	)

	fs, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	err = jsoniter.NewDecoder(fs).Decode(&Config)
	if err != nil && err != io.EOF {
		panic(err)
	}
}
