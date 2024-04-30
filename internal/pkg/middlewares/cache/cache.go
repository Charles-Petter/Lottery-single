package cache

import (
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"lottery_single/internal/pkg/middlewares/log"
	"strconv"
	"time"
)

var (
	client    = &Client{}
	redisConn *redis.Client
)

type Client struct {
}

type Options struct {
	passWord string
	db       int
	poolSize int
	addr     string // 地址，IP:PORT
}

type Option func(*Options)

func WithPassWord(passWord string) Option {
	return func(o *Options) {
		o.passWord = passWord
	}
}

func WithDB(db int) Option {
	return func(o *Options) {
		o.db = db
	}
}

func WithPoolSize(poolSize int) Option {
	return func(o *Options) {
		o.poolSize = poolSize
	}
}

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.addr = addr
	}
}
func newOptions(opts ...Option) Options {
	// 默认配置
	options := Options{
		addr:     "127.0.0.1:6379",
		db:       0,
		poolSize: 10,
		passWord: "",
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// initRedis 初始化db连接
func Init(options ...Option) {
	newRedisConn(newOptions(options...))
}

func newRedisConn(options Options) {
	redisConn = redis.NewClient(&redis.Options{
		Addr:     options.addr,
		Password: options.passWord,
		DB:       options.db,
		PoolSize: options.poolSize,
	})
	if redisConn == nil {
		panic("failed to call redis.NewClient")
	}
	_, err := redisConn.Ping(context.Background()).Result()
	if err != nil {
		panic("Failed to ping redis, err:%s")
	}
}

func (client *Client) Close() {
	if redisConn != nil {
		redisConn.Close()
	}
}

// GetRedisCli 获取数据库连接
func GetRedisCli() *Client {
	return client
}

// IncrBy: val + count
func (client *Client) IncrBy(ctx context.Context, key string, count int64) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()

	temp, err := conn.IncrBy(ctx, key, count).Result()
	if err != nil {
		log.Errorf("Redis IncrBy error: %v", err.Error())
		return 0, err
	}
	return temp, nil
}

// Incr: val++
func (client *Client) Incr(ctx context.Context, key string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()

	temp, err := conn.Incr(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis IncrBy error: %v", err.Error())
		return 0, err
	}
	return temp, nil
}

// Decr: val--
func (client *Client) Decr(ctx context.Context, key string) (string, error) {

	conn := redisConn.Conn()
	defer conn.Close()

	temp, err := conn.Decr(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis Decr error: %v", err.Error())
		return "", err
	}
	ret := strconv.FormatInt(temp, 10)
	return ret, nil
}

// DescBy: val-count
func (client *Client) DecrBy(ctx context.Context, key string, count int64) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()

	temp, err := conn.DecrBy(ctx, key, count).Result()
	if err != nil {
		log.Errorf("Redis DecrBy error: %v", err.Error())
		return 0, err
	}
	return temp, nil
}

// Set: set key value expireTime
func (client *Client) Set(ctx context.Context, key, value string, expireTime time.Duration) error {
	conn := redisConn.Conn()
	defer conn.Close()

	_, err := conn.Set(ctx, key, value, expireTime).Result()
	if err != nil {
		log.Errorf("Redis Set error: %v", err.Error())
		return err
	}

	return nil
}

// SetNX a key/value 已存在返回错误信息（true为设置成功，false为设置失败）
func (client *Client) SetNX(ctx context.Context, key, value string, expireTime time.Duration) (bool, error) {
	conn := redisConn.Conn()
	defer conn.Close()

	ret, err := conn.SetNX(ctx, key, value, expireTime).Result()
	if err != nil {
		log.Errorf("Redis SetNX error: %v", err.Error())
		return ret, err
	}

	return ret, nil
}

// Exists 检查key是否存在
func (client *Client) Exists(ctx context.Context, key string) bool {
	conn := redisConn.Conn()
	defer conn.Close()

	exists, err := conn.Exists(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis Exists error: %v", err.Error())
		return false
	}

	if exists == 1 {
		return true
	}

	return false
}

// Eval 执行lua脚本（执行成功返回true，执行失败返回false）
func (client *Client) EvalBool(ctx context.Context, script string, keys []string, args ...interface{}) (bool, error) {
	conn := redisConn.Conn()
	defer conn.Close()

	ret, err := conn.Eval(ctx, script, keys, args...).Bool()
	if err != nil {
		log.Errorf("Redis Eval error: %v", err.Error())
		return ret, err
	}
	return ret, nil
}

// Eval 执行lua脚本（返回所有执行结果的返回值，用切片组装
func (client *Client) EvalResults(ctx context.Context, script string, keys []string, args ...interface{}) ([]interface{}, error) {
	conn := redisConn.Conn()
	defer conn.Close()

	ret, err := conn.Eval(ctx, script, keys, args...).Slice()
	if err != nil {
		log.Errorf("Redis Eval error: %v", err.Error())
		return ret, err
	}
	return ret, nil
}

// Get: 如果key不存在，那么value返回""，error为nil，true表示key存在，false表示不存在
func (client *Client) Get(ctx context.Context, key string) (string, bool, error) {
	conn := redisConn.Conn()
	defer conn.Close()

	ret, err := conn.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", false, nil
	}

	if err != nil {
		log.Errorf("Redis Get error: %v", err.Error())
		return "", false, err
	}

	return ret, true, nil
}

// Delete 删除一个key
func (client *Client) Delete(ctx context.Context, key string) error {
	conn := redisConn.Conn()
	defer conn.Close()

	_, err := conn.Del(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis Delete error: %v", err.Error())
		return err
	}

	return nil
}

// TTL get the key remain expire time(-2: 表示key不存在 / -1表示key没有过期时间)
func (client *Client) TTL(ctx context.Context, key string) (int, error) {
	conn := redisConn.Conn()
	defer conn.Close()

	ret, err := conn.TTL(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis TTL error: %v", err.Error())
		return -3, err
	}

	switch ret {
	case -2:
		log.Infof("get TTL key is no exist")
		return -2, nil
	case -1:
		log.Infof("get TTL key is no expireTime")
		return -1, nil
	default:
		return int(ret.Seconds()), nil
	}
}

// Expire True: key存在，设置过期时间成功 / False: key不存在，设置过期时间失败
func (client *Client) Expire(ctx context.Context, key string, expire time.Duration) bool {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.Expire(ctx, key, expire).Result()
	if err != nil {
		log.Errorf("Redis Expire error: %v", err.Error())
		return false
	}
	return ret
}

/*
以下是关于对Set数据类型的操作
*/

// SAdd 添加元素到Set中，返回添加成功的个数
func (client *Client) SAdd(ctx context.Context, key string, value ...string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.SAdd(ctx, key, value).Result()
	if err != nil {
		log.Errorf("Redis SAdd error: %v", err.Error())
		return -1, err
	}

	return ret, nil
}

// SRem 移除元素，返回移除成功的个数
func (client *Client) SRem(ctx context.Context, key string, value ...string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.SRem(ctx, key, value).Result()
	if err != nil {
		log.Errorf("Redis SRem error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// SIsMember 判断value是否在set中
func (client *Client) SIsMember(ctx context.Context, key string, value string) (bool, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.SIsMember(ctx, key, value).Result()
	if err != nil {
		log.Errorf("Redis SISMEMBER error: %v", err.Error())
		return false, err
	}
	return ret, nil
}

// SMembers 返回对应set 所有元素
func (client *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.SMembers(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis SMembers error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// SInter 返回交集，返回多个set的交集
func (client *Client) SInter(ctx context.Context, key ...string) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.SInter(ctx, key...).Result()
	if err != nil {
		log.Errorf("Redis SInter error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// SUnion 返回并集，返回多个set的并集
func (client *Client) SUnion(ctx context.Context, key ...string) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.SUnion(ctx, key...).Result()
	if err != nil {
		log.Errorf("Redis SUnion error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// SDiff 返回差集，返回多个set的差集
func (client *Client) SDiff(ctx context.Context, key ...string) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.SDiff(ctx, key...).Result()
	if err != nil {
		log.Errorf("Redis SDiff error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// SCard 返回set中元素的个数
func (client *Client) SCard(ctx context.Context, key string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.SCard(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis SCard error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// SPop 从set中弹出一个元素
func (client *Client) SPop(ctx context.Context, key string) (string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.SPop(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis SPop error: %v", err.Error())
		return "", err
	}
	return ret, nil
}

// SScan 扫描set中的元素
func (client *Client) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, cursor, err := conn.SScan(ctx, key, cursor, match, count).Result()
	if err != nil {
		log.Errorf("Redis SScan error: %v", err.Error())
		return nil, cursor, err
	}
	return ret, cursor, nil
}

/*
以下是关于对Hash数据类型的操作
*/
// HGet 获取hash key对应的value，返回获取成功的value
func (client *Client) HGet(ctx context.Context, key, field string) (string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.HGet(ctx, key, field).Result()
	if err != nil {
		log.Errorf("Redis HGet error: %v", err.Error())
		return "", err
	}
	return ret, nil
}

// HSet 设置hash key对应的value，返回设置成功的个数
func (client *Client) HSet(ctx context.Context, key, field string, value interface{}) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.HSet(ctx, key, field, value).Result()
	if err != nil {
		log.Errorf("Redis HSet error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// HMSet 设置hash key对应的value
func (client *Client) HMSet(ctx context.Context, key string, value map[string]interface{}) (bool, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.HMSet(ctx, key, value).Result()
	if err != nil {
		log.Errorf("Redis HMSet error: %v", err.Error())
		return false, err
	}
	return ret, nil
}

func (client *Client) HIncrBy(ctx context.Context, key, field string, value int64) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	count, err := conn.HIncrBy(ctx, key, field, value).Result()
	if err != nil {
		log.Errorf("Redis HIncrBy error: %v", err.Error())
		return 0, err
	}
	return count, nil
}

// HMGet 获取hash key对应的value，返回获取成功的value
func (client *Client) HMGet(ctx context.Context, key string, fields ...string) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.HMGet(ctx, key, fields...).Result()
	if err != nil {
		log.Errorf("Redis HMGet error: %v", err.Error())
		return nil, err
	}

	retSlice := make([]string, len(ret))
	// []interface{} 转 []string
	for i, v := range ret {
		retSlice[i] = v.(string)
	}

	return retSlice, nil
}

// HKeys 获取hash key对应的member，返回获取成功的member
func (client *Client) HKeys(ctx context.Context, key string) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.HKeys(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis HKeys error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// HLen 获取hash key对应的member数量
func (client *Client) HLen(ctx context.Context, key string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.HLen(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis HLen error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// HDel 删除hash key对应的value
func (client *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.HDel(ctx, key, fields...).Result()
	if err != nil {
		log.Errorf("Redis HDel error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// HExists 判断hash key对应的value是否存在
func (client *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.HExists(ctx, key, field).Result()
	if err != nil {
		log.Errorf("Redis HExists error: %v", err.Error())
		return false, err
	}
	return ret, nil
}

// HGetAll 获取hash key对应的value，返回获取成功的value
func (client *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.HGetAll(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis HGetAll error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// HScan 扫描hash key对应的value，返回扫描成功的value
func (client *Client) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) (map[string]string, uint64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, cursor, err := conn.HScan(ctx, key, cursor, match, count).Result()
	if err != nil {
		log.Errorf("Redis HScan error: %v", err.Error())
		return nil, cursor, err
	}

	// 转换成map
	retMap := make(map[string]string)

	for i := 0; i < len(ret); i += 2 {
		retMap[ret[i]] = ret[i+1]
	}

	return retMap, cursor, nil
}

/*
以下是关于对List数据类型的操作
*/

// LPush 添加元素到List中，返回添加成功的个数
func (client *Client) LPush(ctx context.Context, key string, value ...string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.LPush(ctx, key, value).Result()
	if err != nil {
		log.Errorf("Redis LPush error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// RPush 添加元素到List中，返回添加成功的个数
func (client *Client) RPush(ctx context.Context, key string, value ...string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.RPush(ctx, key, value).Result()
	if err != nil {
		log.Errorf("Redis RPush error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// LPop 移除List中的第一个元素，返回移除成功的元素
func (client *Client) LPop(ctx context.Context, key string) (string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.LPop(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis LPop error: %v", err.Error())
		return "", err
	}
	return ret, nil
}

// RPop 移除List中的最后一个元素，返回移除成功的元素
func (client *Client) RPop(ctx context.Context, key string) (string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.RPop(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis RPop error: %v", err.Error())
		return "", err
	}
	return ret, nil
}

// LLen 获取List中元素的个数，返回获取成功的个数
func (client *Client) LLen(ctx context.Context, key string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.LLen(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis LLen error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// LTrim 让列表只保留指定区间内的元素，不在指定区间之内的元素都将被删除，Result返回OK
func (client *Client) LTrim(ctx context.Context, key string, start, stop int64) (string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.LTrim(ctx, key, start, stop).Result()
	if err != nil {
		log.Errorf("Redis LTrim error: %v", err.Error())
		return "", err
	}
	return ret, nil
}

// LRange 获取List中指定区间内的元素，返回获取成功的元素
func (client *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.LRange(ctx, key, start, stop).Result()
	if err != nil {
		log.Errorf("Redis LRange error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// LIndex 获取List中指定位置的元素，返回获取成功的元素
func (client *Client) LIndex(ctx context.Context, key string, index int64) (string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.LIndex(ctx, key, index).Result()
	if err != nil {
		log.Errorf("Redis LIndex error: %v", err.Error())
		return "", err
	}
	return ret, nil
}

/*
以下是关于对ZSet数据类型的操作
*/
// ZAdd 添加元素到ZSet中，返回添加成功的个数
func (client *Client) ZAdd(ctx context.Context, key string, score float64, member string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Result()
	if err != nil {
		log.Errorf("Redis ZAdd error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// ZRem 移除元素，返回移除成功的个数
func (client *Client) ZRem(ctx context.Context, key string, member string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZRem(ctx, key, member).Result()
	if err != nil {
		log.Errorf("Redis ZRem error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// ZCard 获取ZSet中元素的个数，返回获取成功的个数
func (client *Client) ZCard(ctx context.Context, key string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZCard(ctx, key).Result()
	if err != nil {
		log.Errorf("Redis ZCard error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// ZRange 获取ZSet中指定区间内的元素（按照Rank），返回获取成功的元素
func (client *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		log.Errorf("Redis ZRange error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// ZRevRange 获取ZSet中指定区间内的元素（逆序返回这个区间元素），返回获取成功的元素
func (client *Client) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZRevRange(ctx, key, start, stop).Result()
	if err != nil {
		log.Errorf("Redis ZRevRange error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// ZScore 获取ZSet中指定元素的score，返回获取成功的score
func (client *Client) ZScore(ctx context.Context, key string, member string) (float64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZScore(ctx, key, member).Result()
	if err != nil {
		log.Errorf("Redis ZScore error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// ZRank 获取ZSet中指定元素的rank，返回获取成功的rank
func (client *Client) ZRank(ctx context.Context, key string, member string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZRank(ctx, key, member).Result()
	if err != nil {
		log.Errorf("Redis ZRank error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// ZRevRank 获取ZSet中指定元素的rank，返回获取成功的rank
func (client *Client) ZRevRank(ctx context.Context, key string, member string) (int64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZRevRank(ctx, key, member).Result()
	if err != nil {
		log.Errorf("Redis ZRevRank error: %v", err.Error())
		return -1, err
	}
	return ret, nil
}

// ZRangeByScore 获取ZSet中指定区间内的元素（按照Score），返回获取成功的元素
func (client *Client) ZRangeByScore(ctx context.Context, key string, start, stop string) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: start,
		Max: stop,
	}).Result()
	if err != nil {
		log.Errorf("Redis ZRangeByScore error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// ZRevRangeByScore 获取ZSet中指定区间内的元素（逆序返回这个区间元素），返回获取成功的元素
func (client *Client) ZRevRangeByScore(ctx context.Context, key string, start, stop string) ([]string, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, err := conn.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Max: start,
		Min: stop,
	}).Result()
	if err != nil {
		log.Errorf("Redis ZRevRangeByScore error: %v", err.Error())
		return nil, err
	}
	return ret, nil
}

// ZScan 扫描ZSet中的元素，返回扫描成功的元素
func (client *Client) ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) (map[string]string, uint64, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	ret, cursor, err := conn.ZScan(ctx, key, cursor, match, count).Result()
	if err != nil {
		log.Errorf("Redis ZScan error: %v", err.Error())
		return nil, 0, err
	}

	// 转换成map
	retMap := make(map[string]string)
	for i := 0; i < len(ret); i += 2 {
		retMap[ret[i]] = ret[i+1]
	}

	return retMap, cursor, nil
}

// Rename 重命名key
func (client *Client) Rename(ctx context.Context, key, newKey string) (bool, error) {
	conn := redisConn.Conn()
	defer conn.Close()
	_, err := conn.Rename(ctx, key, newKey).Result()
	if err != nil {
		log.Errorf("Rename ZScan error: %v", err.Error())
		return false, err
	}
	return true, nil
}
