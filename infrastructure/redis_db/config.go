package redis_db

//var Config RedisConfig

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func (r *RedisConfig) Load() {
	r.Addr = "localhost:6379"
	r.Password = ""
	//r.Addr = os.Getenv("Addr")
	//r.Password = os.Getenv("Password")
	//DB, err := strconv.ParseInt(os.Getenv("DB"), 10, 32)
	//if err != nil {
	//	panic(err)
	//}
	//r.DB = int(DB)
	r.DB = 0
}
