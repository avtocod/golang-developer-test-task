package redclient

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestNewRedisClientPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	config := RedisConfig{Addr: ":", Password: "", DB: 0}
	_ = NewRedisClient(context.Background(), config)
}

func TestNewRedisClient(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("the code did panic")
		}
	}()

	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	config := RedisConfig{Addr: mr.Addr(), Password: "", DB: 0}
	_ = NewRedisClient(context.Background(), config)
}
