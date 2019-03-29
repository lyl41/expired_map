package main

import (
	emap "expiredMap/expired_map"
	"fmt"
	"time"
)

func main () {
	cache := emap.NewExpiredMap()

	cache.Set(1, 1, 10)
	fmt.Println(cache.Get(1))
	time.Sleep(time.Second)
	fmt.Println(cache.Get(1))
	fmt.Println(cache.TTL(1))

	cache.Delete(1)
	fmt.Println(cache.Get(1))

	fmt.Println(cache.TTL(1))


	time.Sleep(time.Millisecond * 10)

	for i := 1; i <= 10; i++ {
		cache.Set(i, i, int64(i))
	}
	for i := 1; i <=10; i++ {
		cache.DoForEach(foreach)
		fmt.Println("==========")
		time.Sleep(time.Second)
	}

}

func foreach(key interface{}, val interface{}) {
	fmt.Println("key", key, "val", val)
}


