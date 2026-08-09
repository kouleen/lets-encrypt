package util

import (
	"sync"
	"time"
)

// ExpireMap 自动过期的 KV 存储
type ExpireMap struct {
	items map[string]*item
	lock  sync.RWMutex
}

// item 存储结构 + 过期时间
// 兼容普通字符串缓存 + Incr数字自增
type item struct {
	strVal string // 普通Set字符串值
	numVal int64  // Incr自增数字
	isNum  bool   // true=数字计数器 false=普通字符串
	expire time.Time
}

//// CacheStore 全局实例（项目中直接用这个）
//var CacheStore = NewExpireMap()

var cacheStore = newExpireMap()

func GetCacheStore() *ExpireMap {
	return cacheStore
}

func newExpireMap() *ExpireMap {
	em := &ExpireMap{
		items: make(map[string]*item),
	}
	// 启动后台清理协程，每 10 秒清理一次过期数据
	go em.cleanLoop(10 * time.Second)
	return em
}

//// NewExpireMap 创建自动过期 map
//func NewExpireMap() *ExpireMap {
//	em := &ExpireMap{
//		items: make(map[string]*item),
//	}
//	// 启动后台清理协程，每 10 秒清理一次过期数据
//	go em.cleanLoop(10 * time.Second)
//	return em
//}

// Set 设置键值对 + 过期时间（比如 5 分钟）
func (e *ExpireMap) Set(key, value string, ttl time.Duration) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if key == "" {
		return
	}
	e.items[key] = &item{
		strVal: value,
		isNum:  false,
		expire: time.Now().Add(ttl),
	}
}

// Get 获取值，如果不存在或已过期返回 ""
func (e *ExpireMap) Get(key string) string {
	e.lock.RLock()
	defer e.lock.RUnlock()
	if key == "" {
		return ""
	}
	it, ok := e.items[key]
	if !ok {
		return ""
	}
	// 已过期 增加惰性删除
	if time.Now().After(it.expire) {
		// RLock 不能执行delete，延迟到下一次定时清理
		return ""
	}
	if it.isNum {
		return ""
	}
	return it.strVal
}

// Ttl 查询剩余过期时间，不存在/过期返回 -2
func (e *ExpireMap) Ttl(key string) time.Duration {
	e.lock.RLock()
	defer e.lock.RUnlock()
	if key == "" {
		return -2
	}
	it, ok := e.items[key]
	if !ok {
		return -2
	}
	now := time.Now()
	if now.After(it.expire) {
		return -2
	}
	return it.expire.Sub(now)
}

// Expire 给已存在key设置过期时间（Redis同款）
func (e *ExpireMap) Expire(key string, ttl time.Duration) bool {
	e.lock.Lock()
	defer e.lock.Unlock()
	if key == "" {
		return false
	}
	it, ok := e.items[key]
	if !ok {
		return false
	}
	now := time.Now()
	if now.After(it.expire) {
		delete(e.items, key)
		return false
	}
	it.expire = now.Add(ttl)
	return true
}

// Delete 删除 key
func (e *ExpireMap) Delete(key string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if key == "" {
		return
	}
	delete(e.items, key)
}

// IntResult 模仿redis.IntCmd返回结构，兼容 .Result() 调用
type IntResult struct {
	val int64
	err error
}

func (ir *IntResult) Result() (int64, error) {
	return ir.val, ir.err
}

// Incr 自增计数器，对齐Redis INCR
// key不存在/已过期：创建，值为1
// key存在数字类型：+1
// key存在字符串类型：返回错误
func (e *ExpireMap) Incr(key string) *IntResult {
	e.lock.Lock()
	defer e.lock.Unlock()
	if key == "" {
		return &IntResult{err: nil}
	}
	now := time.Now()
	it, exists := e.items[key]

	// 不存在 或 已过期 → 新建计数器，初始值1
	if !exists || now.After(it.expire) {
		e.items[key] = &item{
			numVal: 1,
			isNum:  true,
			expire: time.Now().Add(time.Duration(10) * time.Minute),
		}
		return &IntResult{val: 1, err: nil}
	}

	// 原有key是普通字符串，不支持自增
	if !it.isNum {
		return &IntResult{err: nil}
	}

	// 数字自增
	it.numVal++
	return &IntResult{val: it.numVal, err: nil}
}

// IncrBy 自定义步长自增
func (e *ExpireMap) IncrBy(key string, step int64) *IntResult {
	e.lock.Lock()
	defer e.lock.Unlock()
	if key == "" {
		return &IntResult{err: nil}
	}
	now := time.Now()
	it, exists := e.items[key]

	if !exists || now.After(it.expire) {
		e.items[key] = &item{
			numVal: step,
			isNum:  true,
			expire: time.Time{},
		}
		return &IntResult{val: step, err: nil}
	}

	if !it.isNum {
		return &IntResult{err: nil}
	}

	it.numVal += step
	return &IntResult{val: it.numVal, err: nil}
}

// GetNum 获取计数器当前值，不存在/过期/非数字返回0
func (e *ExpireMap) GetNum(key string) int64 {
	e.lock.RLock()
	defer e.lock.RUnlock()
	if key == "" {
		return 0
	}
	it, ok := e.items[key]
	if !ok || time.Now().After(it.expire) || !it.isNum {
		return 0
	}
	return it.numVal
}

// 后台自动清理过期数据
func (e *ExpireMap) cleanLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		e.clean()
	}
}

// 执行清理
func (e *ExpireMap) clean() {
	e.lock.Lock()
	defer e.lock.Unlock()
	now := time.Now()
	for key, it := range e.items {
		if now.After(it.expire) {
			delete(e.items, key)
		}
	}
}
