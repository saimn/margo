package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	sum := 0
	for _, arg := range os.Args[1:] {
		// fmt.Printf("%d: %q %t\n", i, arg, arg)
		v, err := strconv.Atoi(arg)

		if err != nil {
			fmt.Fprintf(os.Stderr, "oups... %v\n", err)
			panic(err)
			// os.Exit(1)
		}
		sum += v
	}
	fmt.Printf("sum=%d\n", sum)
}
