package expire_map

import (
	"fmt"
	"testing"
	"time"
)

func Test_Map(t *testing.T) {
	cache := NewExpiredMap()

	for i := 1; i <= 10; i++ {
		cache.Set(i, i, int64(i))
	}
	cache.Delete(8)
	cache.Delete(9)
	cache.Delete(10)
	for i := 1; i <= 100; i++ {
		cache.Set(10+i, 8, 5)
	}
	for i := 1; i <= 10; i++ {
		cache.DoForEach(foreach)
		fmt.Println("==========")
		time.Sleep(time.Second)
	}
	cache.Close()
	time.Sleep(time.Millisecond)
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Millisecond * 1)
		cache.Set(i, i, int64(i))
		fmt.Println(cache.Get(i))
	}
}

func foreach(key interface{}, val interface{}) {
	k := key.(int)
	if k > 10 {
		return
	}
	fmt.Println("key", k, "val", val)
}
