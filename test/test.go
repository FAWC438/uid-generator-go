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
