package main

import (
	"fmt"
	"sync"
	"time"
)

type val struct {
	data interface{}
	expiredTime int64
}

type ExpiredMap struct {
	m map[interface{}]*val
	timeMap map[int64][]interface{}
	lck *sync.Mutex
	timeMapLck *sync.Mutex
	stop chan bool
}

func NewExpiredMap() (*ExpiredMap) {
	e := ExpiredMap{
		m : make(map[interface{}]*val),
		lck : new(sync.Mutex),
		timeMap: make(map[int64][]interface{}),
		timeMapLck: new(sync.Mutex),
		stop : make(chan bool),
	}
	go e.run()
	return &e
}
//background goroutine 主动删除过期的key
//因为删除需要花费时间，这里启动goroutine来删除，但是启动goroutine和now++也需要少量时间，
//导致数据实际删除时间比应该删除的时间稍晚一些，这个误差我们应该能接受。
func (e *ExpiredMap) run() {
	now := time.Now().Unix()
	t := time.NewTicker(time.Second)
	for {
		select {
		case <- t.C:
			now++   //这里用now++的形式，直接用time.Now().Unix()可能会导致时间跳过，导致key未删除。
			//fmt.Println("now: ", now, "realNow", time.Now().Unix())
			go func(now int64) {
				if keys, found := e.timeMap[now]; found { //todo delete timeMap
					e.MultiDelete(keys)  //todo 应该是list
					e.timeMapLck.Lock()
					defer e.timeMapLck.Unlock()
					delete(e.timeMap, now)
				}
			}(now)
		}
		select {  //不放在同一个select中，防止同时收到两个信号后随机选择导致没有return
		case <- e.stop:
			fmt.Println("=== STOP ===")
			return
		default:
		}
	}
}

func (e *ExpiredMap) Set(key, value interface{}) {
	e.lck.Lock()
	defer e.lck.Unlock()
	e.m[key] = &val{
		data: value,
		expiredTime: -1,
	}
}

func (e *ExpiredMap) SetWithExpired(key, value interface{}, expiredSeconds int64){
	if expiredSeconds <= 0 {
		return
	}
	e.lck.Lock()
	defer e.lck.Unlock()
	expiredTime := time.Now().Unix() + expiredSeconds
	e.m[key] = &val{
		data:        value,
		expiredTime: expiredTime,
	}
	e.timeMapLck.Lock()
	defer e.timeMapLck.Unlock()
	if keys, found := e.timeMap[expiredTime]; found {
		keys = append(keys, key)
		e.timeMap[expiredTime] = keys
	} else {
		keys = append(keys, key)
		e.timeMap[expiredTime] = keys
	}
}

//
func (e *ExpiredMap) Get(key interface{}) (interface{}){
	e.lck.Lock()
	defer e.lck.Unlock()
	if value, found := e.m[key]; found {
		return value.data
	}
	return nil
}

func (e *ExpiredMap) Delete(key interface{}) {
	e.lck.Lock()
	defer e.lck.Unlock()
	delete(e.m, key)
}

func (e *ExpiredMap) MultiDelete(keys []interface{}) {
	e.lck.Lock()
	defer e.lck.Unlock()
	for _, key := range keys {
		if _, found := e.m[key]; found {
			delete(e.m, key)
		}
	}
}

func (e *ExpiredMap) Length() int {
	e.lck.Lock()
	defer e.lck.Unlock()
	return len(e.m)
}

func (e *ExpiredMap) Size() int {
	return e.Length()
}

//返回key的剩余生存时间 key不存在返回-2，key没有设置生存时间返回-1
func (e *ExpiredMap) TTL (key interface{}) int64 {
	e.lck.Lock()
	defer e.lck.Unlock()
	if value, found := e.m[key]; found {
		if value.expiredTime == -1 {
			return -1
		}
		now := time.Now().Unix()
		if value.expiredTime - now < 0 {
			go e.Delete(key)
			return -2
		}
		return value.expiredTime - now
	} else {
		return -2
	}
}

func (e *ExpiredMap) Clear() {
	e.lck.Lock()
	defer e.lck.Unlock()
	e.m = make(map[interface{}]*val)
	e.timeMap = make(map[int64][]interface{})
}

func (e *ExpiredMap) Close () {
	e.lck.Lock()
	defer e.lck.Unlock()
	e.m = nil
	e.timeMap = nil
	e.stop <- true
}

func (e *ExpiredMap) Stop () {
	e.Close()
}

func (e *ExpiredMap) DoForEach(handler func (interface{}, *val)) {
	e.lck.Lock()
	defer e.lck.Unlock()
	for k,v := range e.m {
		handler(k, v)
	}
}

func print(key interface{}, value *val) {
	fmt.Println("key:", key, "value:", value)
}

func main () {
	redis := NewExpiredMap()

	redis.Set(1, 2)
	fmt.Println("len-1:", redis.Length())
	redis.DoForEach(print)

	redis.SetWithExpired(2, 3, 2)
	redis.SetWithExpired(3, 4, 2)
	redis.DoForEach(print)
	fmt.Println("len-3:", redis.Length())
	time.Sleep(time.Millisecond*1000)
	fmt.Println("TTL-1:", redis.TTL(2))
	time.Sleep(time.Millisecond*1010)
	fmt.Println("len-1:", redis.Length())

	redis.SetWithExpired(3, 4, 3)
	fmt.Println("len-2:", redis.Length())


	redis.Stop()
}

