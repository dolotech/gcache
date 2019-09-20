package internal

import (
	"encoding/json"
	"github.com/8treenet/gcache/option"
	"github.com/go-redis/redis"
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type TestEmail struct {
	gorm.Model
	TypeID     int
	Subscribed bool
	TestUserID int
}

var modelValue string = `
{\"PK\":\"18\",\"Model\":{\"ID\":18,\"CreatedAt\":\"2019-09-15T16:30:16+08:00\",\"UpdatedAt\":\"2019-09-15T16:30:16+08:00\",\"DeletedAt\":null,\"TypeID\":18,\"Subscribed\":false,\"TestUserID\":18}}
`

func TestInitRefreshData(t *testing.T) {
	cp := gettestcachePlugin()
	cp.FlushDB()
	cp.handle.redisClient.Set("test_emails_model_18", modelValue, 180*time.Second).Err()
	var js JsonSearch
	js.Primarys = append(js.Primarys, 18)
	js.UpdatedAt = time.Now().Unix() - 5000000
	buffer, _ := json.Marshal(js)
	cp.handle.redisClient.HSet("test_emails_search_&((type_id>=$)", "18_LIMIT_1", buffer)
	cp.handle.redisClient.Expire("test_emails_search_&((type_id>=$)", 180*time.Second)

	cp.handle.redisClient.HSet("test_emails_affect_type_id", "test_emails_search_&((type_id>=$)", js.UpdatedAt)
}

func TestRefresh(t *testing.T) {
	cp := gettestcachePlugin()
	cp.Debug()

	dh := cp.handle.NewDeleteHandle()
	dh.refresh(reflect.TypeOf(TestEmail{}))
}

func gettestcachePlugin() *plugin {
	addr := "root:123123@tcp(127.0.0.1:3306)/matrix?charset=utf8&parseTime=True&loc=Local"
	db, e := gorm.Open("mysql", addr)
	if e != nil {
		panic(e)
	}

	opt := option.DefaultOption{}
	//缓存插件 注入到Gorm。开启Debug，查看日志
	cachePlugin := InjectGorm(db, &opt, &option.RedisOption{Addr:"localhost:6379"})
	return cachePlugin
}

func TestLuaRefresh(t *testing.T) {
	c := gettestcachePlugin()
	script := redis.NewScript(`
	local all = redis.call("HGETALL", KEYS[1])
	redis.log(redis.LOG_NOTICE, #all);
	for k,v in pairs(all) do
		redis.log(redis.LOG_NOTICE, k);
		redis.log(redis.LOG_NOTICE, v);
	end
	return true
`)
	script.Run(c.handle.redisClient, []string{"test_users_search_&((test_users.user_name=$)&(test_users.age=$))"}, 1).Result()
}