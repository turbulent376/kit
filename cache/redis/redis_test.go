//go:build integration
// +build integration

package redis

import (
	"fmt"
	"git.jetbrains.space/orbi/fcsd/kit/log"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var logger = log.Init(&log.Config{Level: log.TraceLevel})

func Test_Range(t *testing.T) {

	cl, _ := Open(&Config{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		Ttl:      0,
	}, func() log.CLogger {
		return log.L(logger)
	})

	key := "test-list"

	if jsons, err := cl.Instance.LRange(key, 0, -1).Result(); err == nil {
		fmt.Println(jsons)
	} else {
		t.Fatal(err)
	}

	pipe := cl.Instance.Pipeline()
	pipe.Expire(key, time.Second*10)

	if err := cl.Instance.RPush(key, "1").Err(); err != nil {
		t.Fatal(err)
	}
	if err := cl.Instance.RPush(key, "2").Err(); err != nil {
		t.Fatal(err)
	}
	if err := cl.Instance.RPush(key, "3").Err(); err != nil {
		t.Fatal(err)
	}

	_, err := pipe.Exec()
	if err != nil {
		t.Fatal(err)
	}

	if jsons, err := cl.Instance.LRange(key, 0, -1).Result(); err == nil {
		fmt.Println(jsons)
		assert.Equal(t, 3, len(jsons))
	} else {
		t.Fatal(err)
	}

}
