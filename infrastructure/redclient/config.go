package redclient

import (
	"os"
	"strconv"
)

// RedisConfig is struct for storing data about path to Redis Storage
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Load is useful for loading RedisConfig data
func (r *RedisConfig) Load() {
	r.Addr = os.Getenv("Addr")
	r.Password = os.Getenv("Password")
	DB, err := strconv.ParseInt(os.Getenv("DB"), 10, 32)
	if err != nil {
		panic(err)
	}
	r.DB = int(DB)
}
