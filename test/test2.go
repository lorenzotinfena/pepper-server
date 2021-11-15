package main

import (
	"context"
	"fmt"
)

/*func main(){
	c := make(chan bool)
	go func(){
		select {
		case <-c:
			fmt.Println("return")
			return
		default:
			time.Sleep(3*time.Second)
			fmt.Println(3)
		}
	}()
	time.Sleep(6*time.Second)
	c<-true
}*/
func gen(ctx context.Context) <-chan int {
	ch := make(chan int)
	go func() {
		var n int
		for {
			select {
			case <-ctx.Done():
				return
			case ch <- n:
				n++
			}
		}
	}()
	return ch
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for n := range gen(ctx) {
		fmt.Println(n)
		if n == 5 {
			cancel()
			break
		}
	}
}
