package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

// Sqrt compute the square root, yay!
func Sqrt(x float64, niter int) float64 {
	zn := 1.0
	for i := 0; i < niter; i++ {
		zn = zn - (zn*zn-x)/(2*zn)
	}
	return zn
}

func main() {
	v, err := strconv.ParseFloat(os.Args[1], 64)
	precision := 10
	if len(os.Args) > 2 {
		precision, _ = strconv.Atoi(os.Args[2])
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "oups... %v\n", err)
		panic(err)
	}

	fmt.Printf("Newton: %v\n", Sqrt(v, precision))
	fmt.Printf("Math: %v\n", math.Sqrt(v))
	fmt.Printf("Diff: %v\n", math.Sqrt(v)-Sqrt(v, precision))
}
