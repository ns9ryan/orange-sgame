package redis_factory

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"time"
	"reflect"
	"strconv"
	"github.com/shopspring/decimal"
	"encoding/json"
)

// RedisStorage 封装 Redis 操作
type RedisStorage struct {
	client   *redis.Client // *redis.Client内部已经实现了连接池与自我熔断等机制
	host     string
	port     string
	db       int
	password string
}

func InitRedis(host, port, auth string, indexDb int) (error, *RedisStorage) {

	redisStorage := NewRedisStorage(host, port, auth, indexDb)

	//测试 Redis 连接是否成功
	_, err := redisStorage.client.Ping().Result()

	if err != nil {
		return fmt.Errorf("Failed to connect to Redis: %v", err), nil
	}

	fmt.Printf("redis 全局连接初始化成功\n")

	return nil, redisStorage
}

// NewRedisStorage 创建 RedisStorage 实例
func NewRedisStorage(host, port, password string, db int) *RedisStorage {
	var addr bytes.Buffer
	addr.WriteString(host)
	addr.WriteString(":")
	addr.WriteString(port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr.String(),
		Password: password,
		DB:       db,
	})

	return &RedisStorage{
		client:   client,
		host:     host,
		port:     port,
		db:       db,
		password: password,
	}
}

// Close 关闭 Redis 连接
func (r *RedisStorage) Close() error {
	return r.client.Close()
}

// Set 设置键值对
func (r *RedisStorage) Set(key string, value interface{}, exp time.Duration) error {
	err := r.client.Set(key, value, exp).Err()
	return wrapRedisError("Set", err)
}

// Get 获取值
func (r *RedisStorage) Get(key string) (string, error) {
	res, err := r.client.Get(key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return "", wrapRedisError("Get", errors.New("key not found"))
		}
		return "", wrapRedisError("Get", err)
	}
	return res, nil
}

// Exists 检查 key 是否存在
func (r *RedisStorage) Exists(key string) error {
	exists, err := r.client.Exists(key).Result()
	if exists == 0 {
		return wrapRedisError("Exists", errors.New("key does not exist"))
	}
	return wrapRedisError("Exists", err)
}

// Expire 设置 key 的过期时间
func (r *RedisStorage) Expire(key string, t time.Duration) error {
	ok, err := r.client.Expire(key, t).Result()
	if !ok {
		return wrapRedisError("Expire", errors.New("failed to set expiry"))
	}
	return wrapRedisError("Expire", err)
}

// HSet 设置key->field->value
func (r *RedisStorage) HSet(hash, field, value string) error {
	err := r.client.HSet(hash, field, value).Err()
	return wrapRedisError("HSet", err)
}

// HGet 获取某个key的某个field的value
func (r *RedisStorage) HGet(hash, field string) (string, error) {
	value, err := r.client.HGet(hash, field).Result()
	return value, wrapRedisError("HGet", err)
}

// HDel 删除某个key的某个field
func (r *RedisStorage) HDel(hash, field string) (int64, error) {
	result, err := r.client.HDel(hash, field).Result()
	return result, wrapRedisError("HDel", err)
}

// HMSet 设置整个key-data对象，data里的每个属性-值变为file1->value1; file2->value2
func (r *RedisStorage) HMSet(hash string, data map[string]interface{}) error {
	err := r.client.HMSet(hash, data).Err()
	return wrapRedisError("HMSet", err)
}

// HMGet 获取key里的某些fields的值，返回map[string]string
func (r *RedisStorage) HMGet(hash string, fields ...string) (map[string]string, error) {
	values, err := r.client.HMGet(hash, fields...).Result()
	if err != nil {
		return nil, wrapRedisError("HMGet", err)
	}

	result := make(map[string]string)
	for i, field := range fields {
		if value, ok := values[i].(string); ok {
			result[field] = value
		}
	}

	return result, nil
}

// 传入一个结构体，转为map[string]interface{}后，调用HMSet()存入redis
func StructToRedisMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	// 处理指针并逐层解开其指向的值
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return result
		}
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		key := field.Tag.Get("json")
		if key == "" {
			key = field.Name
		}

		fieldValue := v.Field(i)

		// 处理指针字段（解引用）
		for fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				result[key] = nil
				continue
			}
			fieldValue = fieldValue.Elem()
		}

		switch fieldValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			result[key] = fmt.Sprintf("%d", fieldValue.Interface())

		case reflect.Float32, reflect.Float64:
			result[key] = fmt.Sprintf("%f", fieldValue.Interface())

		case reflect.Bool:
			result[key] = strconv.FormatBool(fieldValue.Bool())

		case reflect.String:
			result[key] = fieldValue.String()

		case reflect.Struct:
			switch fieldValue.Type() {
			case reflect.TypeOf(time.Time{}):
				result[key] = fieldValue.Interface().(time.Time).Format(time.RFC3339)
			case reflect.TypeOf(decimal.Decimal{}):
				result[key] = fieldValue.Interface().(decimal.Decimal).String()
			default:
				jsonVal, _ := json.Marshal(fieldValue.Interface())
				result[key] = string(jsonVal)
			}

		case reflect.Slice, reflect.Array, reflect.Map:
			jsonVal, _ := json.Marshal(fieldValue.Interface())
			result[key] = string(jsonVal)

		default:
			result[key] = fmt.Sprintf("%v", fieldValue.Interface()) // 兜底处理
		}
	}

	return result
}

// LPush 向 List 左侧添加元素
func (r *RedisStorage) LPush(list, value string) error {
	err := r.client.LPush(list, value).Err()
	return wrapRedisError("LPush", err)
}

// RPush 向 List 右侧添加元素
func (r *RedisStorage) RPush(list, value string) error {
	err := r.client.RPush(list, value).Err()
	return wrapRedisError("RPush", err)
}

// LPop 从 List 左侧弹出元素
func (r *RedisStorage) LPop(list string) (string, error) {
	value, err := r.client.LPop(list).Result()
	return value, wrapRedisError("LPop", err)
}

// LRem 删除 List 中的元素
func (r *RedisStorage) LRem(list, value string) (int64, error) {
	count, err := r.client.LRem(list, 0, value).Result()
	return count, wrapRedisError("LRem", err)
}

// SAdd 向 Set 添加元素
func (r *RedisStorage) SAdd(set, value string) error {
	err := r.client.SAdd(set, value).Err()
	return wrapRedisError("SAdd", err)
}

// SIsMember 检查 Set 中是否包含某个元素
func (r *RedisStorage) SIsMember(set, value string) (bool, error) {
	isMember, err := r.client.SIsMember(set, value).Result()
	return isMember, wrapRedisError("SIsMember", err)
}

// SRem 从 Set 中移除某个元素
func (r *RedisStorage) SRem(set, value string) (int64, error) {
	count, err := r.client.SRem(set, value).Result()
	return count, wrapRedisError("SRem", err)
}

// GetTime 获取 Redis 服务器时间
func (r *RedisStorage) GetTime() (time.Time, error) {
	t, err := r.client.Time().Result()
	return t, wrapRedisError("GetTime", err)
}

// wrapRedisError 用于包装 Redis 错误信息
func wrapRedisError(functionName string, err error) error {
	if err == nil {
		return nil
	}
	return errors.New("Redis error in " + functionName + ": " + err.Error())
}

// SetNX 原子设置键值对，如果键不存在
func (r *RedisStorage) SetNX(key string, value interface{}, exp time.Duration) (bool, error) {
	ok, err := r.client.SetNX(key, value, exp).Result()
	if err != nil {
		return false, wrapRedisError("SetNX", err)
	}
	return ok, nil
}

// Del 删除键
func (r *RedisStorage) Del(key string) (int64, error) {
	result, err := r.client.Del(key).Result()
	return result, wrapRedisError("Del", err)
}

// Incr 增加指定 key 的值
func (r *RedisStorage) Incr(key string) (int64, error) {
	res, err := r.client.Incr(key).Result()
	return res, wrapRedisError("Incr", err)
}

// IncrBy 增加指定 key 的值 by 指定的数量
func (r *RedisStorage) IncrBy(key string, val int64) (int64, error) {
	res, err := r.client.IncrBy(key, val).Result()
	return res, wrapRedisError("IncrBy", err)
}

// GetGlobalCounter 生成一个基于日期、时间和序列号的全局订单号
func (r *RedisStorage) GetGlobalCounter(prefix string) (string, error) {
	// 获取当前日期和时间
	now := time.Now()
	date := now.Format("20060102")    // YYYYMMDD 格式的日期
	timeStamp := now.Format("150405") // HHMMSS 格式的时间戳

	// Redis 键的命名，用于生成唯一的订单号
	orderIDKey := fmt.Sprintf("order_id_%s_%s", date, timeStamp)

	// 使用 Redis 自增命令生成唯一的订单序列号
	orderSeq, err := r.client.Incr(orderIDKey).Result()
	if err != nil {
		return "", fmt.Errorf("failed to generate order sequence with key %s: %v", orderIDKey, err)
	}

	// 拼接订单号，序列号为 8 位，不足时前补零
	orderNo := fmt.Sprintf("%s-%s%s-%08d", prefix, date, timeStamp, orderSeq)
	return orderNo, nil
}

// Lock 尝试获取分布式锁
func (r *RedisStorage) Lock(lockKey, value string, expiration time.Duration) (bool, error) {

	ok, err := r.client.SetNX(lockKey, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("error trying to acquire lock: %v", err)
	}
	if !ok {
		fmt.Printf("Lock acquisition failed for key: %s\n", lockKey)
	}

	return ok, nil
}

// Unlock 释放分布式锁
func (r *RedisStorage) Unlock(lockKey, value string) error {
	val, err := r.client.Get(lockKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil // 锁已经被释放或从未存在
		}
		return fmt.Errorf("error checking lock: %v", err)
	}

	// 如果锁的值是当前持有的值，则释放锁
	if val == value {
		_, err = r.client.Del(lockKey).Result()
		if err != nil {
			return fmt.Errorf("error releasing lock: %v", err)
		}
	} else {
		return errors.New("lock is not owned by the current process")
	}

	return nil
}
