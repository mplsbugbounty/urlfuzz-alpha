package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(time.Second)
	mil_ticker := time.NewTicker(100*time.Millisecond)
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		time.Sleep(10 * time.Second)
		done <- true
	}()
	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case t := <-ticker.C:
			fmt.Println("SECOND: ", t)
		
		case t := <-mil_ticker.C:
			fmt.Println("MIL: ", t)
		}
	}
}
