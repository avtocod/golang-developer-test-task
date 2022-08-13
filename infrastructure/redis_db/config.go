package redis_db

import (
	"os"
	"strconv"
)

//var Config RedisConfig

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func (r *RedisConfig) Load() {
	r.Addr = os.Getenv("Addr")
	r.Password = os.Getenv("Password")
	DB, err := strconv.ParseInt(os.Getenv("DB"), 10, 32)
	if err != nil {
		panic(err)
	}
	r.DB = int(DB)
}
