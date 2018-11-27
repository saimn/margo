package main

import (
	"fmt"

	"golang.org/x/tour/tree"
)

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *tree.Tree, ch chan int) {
	visit(t, ch)
	close(ch)
}

func visit(t *tree.Tree, ch chan int) {
	if t == nil {
		return
	}
	visit(t.Left, ch)
	ch <- t.Value
	visit(t.Right, ch)
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *tree.Tree) bool {
	ch1 := make(chan int)
	go Walk(t1, ch1)

	ch2 := make(chan int)
	go Walk(t2, ch2)
	for {
		v1, ok1 := <-ch1
		v2, ok2 := <-ch2
		fmt.Printf("%d %v %d %v\n", v1, ok1, v2, ok2)
		if ok1 && ok2 {
			if v1 != v2 {
				return false
			}
		} else if !ok1 && !ok2 {
			return true
		} else {
			return false
		}
	}
}

func main() {
	ch := make(chan int)
	t := tree.New(1)
	fmt.Printf("Tree: %v\n", t)
	go Walk(t, ch)
	for i := range ch {
		fmt.Printf("%d ", i)
	}
	fmt.Printf("\n")

	fmt.Println("same v1")
	ok := Same(t, t)
	if !ok {
		panic(ok)
	}

	fmt.Println("same v2")
	t2 := tree.New(2)
	fmt.Printf("Tree: %v\n", t2)
	ok2 := Same(t, t2)
	if ok2 {
		panic(ok2)
	}

	fmt.Println("same v3")
	t3 := tree.New(1)
	t3.Right = t2
	fmt.Printf("Tree: %v\n", t3)
	ok2 = Same(t, t2)
	if ok2 {
		panic(ok2)
	}
}
