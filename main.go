package main

import (
	"fmt"
	"time"
)


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

