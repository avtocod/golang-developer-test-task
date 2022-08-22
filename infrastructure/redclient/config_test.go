package redclient

import "testing"

func TestRedisConfigLoad(t *testing.T) {
	t.Setenv("Addr", "a")
	t.Setenv("Password", "b")
	t.Setenv("DB", "0")
	config := RedisConfig{}
	config.Load()
}

func TestRedisConfigLoadPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	t.Setenv("Addr", "a")
	t.Setenv("Password", "b")
	// t.Setenv("DB", "0")
	config := RedisConfig{}
	config.Load()
}

func TestRedisConfigLoadPanicConvertDBToInt(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code did not panic")
		}
	}()

	t.Setenv("Addr", "a")
	t.Setenv("Password", "b")
	t.Setenv("DB", "abracadabra")
	config := RedisConfig{}
	config.Load()
}
