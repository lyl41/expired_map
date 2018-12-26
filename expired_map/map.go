package expired_map

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
	alreadyRun bool
	stop chan bool
}

func NewExpiredMap() (*ExpiredMap) {
	e := ExpiredMap{
		m : make(map[interface{}]*val),
		lck : new(sync.Mutex),
		timeMap: make(map[int64][]interface{}),
		stop : make(chan bool),
		alreadyRun: false,
	}
	//go e.run()
	return &e
}
//background goroutine 主动删除过期的key
//因为删除需要花费时间，这里启动goroutine来删除，但是启动goroutine和now++也需要少量时间，
//导致数据实际删除时间比应该删除的时间稍晚一些，这个误差我们应该能接受。
func (e *ExpiredMap) run() {
	now := time.Now().Unix()
	t := time.NewTicker(time.Second)
	del := make(chan []interface{},10)
	go func() {
		for v:=range del{
			e.MultiDelete(v)  //todo 应该是list
		}
	}()
	for {
		select {
		case <- t.C:
			now++   //这里用now++的形式，直接用time.Now().Unix()可能会导致时间跳过，导致key未删除。
			//fmt.Println("now: ", now, "realNow", time.Now().Unix())
			if keys, found := e.timeMap[now]; found { //todo delete timeMap
				del <- keys
			}
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
	if !e.alreadyRun { //lazy:懒启动
		go e.run()
		e.alreadyRun = true
	}
	expiredTime := time.Now().Unix() + expiredSeconds
	e.m[key] = &val{
		data:        value,
		expiredTime: expiredTime,
	}
	//e.timeMapLck.Lock()
	//defer e.timeMapLck.Unlock()
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
	var t int64
	for _, key := range keys {
		if v, found := e.m[key]; found {
			t = v.expiredTime
			delete(e.m, key)
		}
	}
	delete(e.timeMap,t)
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

func (e *ExpiredMap) DoForEach(handler func (interface{}, interface{})) {
	e.lck.Lock()
	defer e.lck.Unlock()
	for k,v := range e.m {
		handler(k, v)
	}
}

func (e *ExpiredMap) DoForEachWithBreak (handler func (interface{}, interface{}) bool) {
	e.lck.Lock()
	defer e.lck.Unlock()
	for k,v := range e.m {
		if handler(k, v) {
			break
		}
	}
}


