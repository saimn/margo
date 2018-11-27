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

func walk(dx int, dy int, nx int, ny int, ch chan bool) {
	// fmt.Printf("%d / %d, %d / %d\n", dx, dy, nx, ny)
	if dx == nx && dy == ny {
		ch <- false // end
		return
	} else if dx == nx {
		walk(dx, dy+1, nx, ny, ch)
	} else if dy == ny {
		walk(dx+1, dy, nx, ny, ch)
	} else {
		ch <- true // start
		go walk(dx, dy+1, nx, ny, ch)
		walk(dx+1, dy, nx, ny, ch)
	}
}

func compute(nx, ny int) int {
	ch := make(chan bool, 100)

	go walk(0, 0, nx, ny, ch)

	ended := 0
	started := 1

	for {
		val := <-ch
		if val {
			started++
		} else {
			ended++
		}
		if ended == started {
			break
		}
	}
	close(ch)
	return ended
}
