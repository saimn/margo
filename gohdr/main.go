package main

import (
	"fmt"
	"os"

	"github.com/astrogo/fitsio"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s FILE\n", os.Args[0])
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1]) // For read access.
	if err != nil {
		panic(err)
	}
	defer file.Close()

	f, err := fitsio.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// get the second HDU
	hdr := f.HDU(0).Header()

	for k := range hdr.Keys() {
		card := hdr.Card(k)
		fmt.Printf(
			"%-8s= %-29s / %s\n",
			card.Name,
			fmt.Sprintf("%v", card.Value),
			card.Comment)
	}
	fmt.Printf("END\n\n")
}
