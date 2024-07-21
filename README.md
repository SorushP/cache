## Concurrent LRU Cache in Go

This package provides a thread-safe implementation of Least Recently Used (LRU) cache. 

### Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/sorushp/cache/pkg/cache"
)

func main() {
	// Create a new cache instance
	myCache := cache.NewCache[string]() // Replace string with your value type

	// Set and Get values
	myCache.Set("key1", "value1", 10*time.Second) // Set key1 with value1 and TTL of 10 seconds
	value, found := myCache.Get("key1")
	if found {
		fmt.Println("Retrieved value: ", value)
	}
}
```

### Test
Read `Makefile` to test and benchmark the package:

```shell
make # all

make test
make coverage
make benchmark
make pprof-cpu
make pprof-mem
make clean
```