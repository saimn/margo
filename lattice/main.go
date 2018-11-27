package main

import (
	"flag"
	"fmt"
)

func main() {
	nn := flag.Int("xy", 2, "NxN size of the grid")
	flag.Parse()

	nx := *nn
	ny := *nn

	fmt.Printf("n=%d\n", compute(nx, ny))
}

func walk(dx int, dy int, nx int, ny int, start chan int, end chan int) {
	// fmt.Printf("%d / %d, %d / %d\n", dx, dy, nx, ny)
	if dx == nx && dy == ny {
		end <- 1
		return
	} else if dx == nx {
		walk(dx, dy+1, nx, ny, start, end)
	} else if dy == ny {
		walk(dx+1, dy, nx, ny, start, end)
	} else {
		go walk(dx+1, dy, nx, ny, start, end)
		start <- 1
		go walk(dx, dy+1, nx, ny, start, end)
	}
}

func compute(nx, ny int) int {
	start := make(chan int)
	end := make(chan int)

	go walk(0, 0, nx, ny, start, end)

	ended := 0
	started := 1

	for {
		select {
		case val := <-start:
			started += val
			// fmt.Printf("started: %d\n", started)
		case val := <-end:
			ended += val
			// fmt.Printf("ended: %d\n", ended)
		}
		if ended == started {
			break
		}
	}
	close(start)
	close(end)
	return ended
}
