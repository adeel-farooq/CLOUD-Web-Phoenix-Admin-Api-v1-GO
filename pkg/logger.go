package pkg

import (
	"fmt"
	"os"
	"strings"
)

var (
	logEnabled bool
)

func init() {
	env := os.Getenv("APP_ENV")
	logEnabled = !strings.EqualFold(env, "production")
}

func Log(args ...interface{}) {
	if logEnabled {
		fmt.Println(args...)
	}
}
