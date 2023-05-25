# Uid-Generator-GO

A Go version of [Uid-Generator](https://github.com/baidu/uid-generator) ,
a [Snowflake](https://github.com/twitter-archive/snowflake) based unique ID generator.

## Status

Work in progress.

- [x] BaseGenerator(DefaultGenerator)
- [ ] CachedGenerator

Welcome to open issues or pull requests to help me complete this project.

## Difference

All code logic is the same as the Java version, but the following differences:

- WorkerIDAssigner is **not implemented**, so you need to assign worker id by yourself.
- Difference between Goroutines and Java Threads
- ...

## Usage

```go
import "github.com/FAWC438/uid-generator-go"

...
```

There is a [demo](/test/test.go) to show how to use `BaseGenerator`.

```go
package test

import (
	"fmt"
	"uid-generator-go/generator"
)

func Test() {
	var gen generator.UidGenerator
	gen = generator.BaseGeneratorConstructor()

	uid, e := gen.GetUID()
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Println(gen.ParseUID(uid))
}
```

Run this test and get your first awesome uid!
