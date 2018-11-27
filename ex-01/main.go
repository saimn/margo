package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("args: %v\n", os.Args)

	for i, arg := range os.Args[1:] {
		fmt.Printf("%d: %q %t\n", i, arg, arg)
	}
}
