package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {

	int_in , err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%v\n" , getBits(int_in))
}

func getBits(n int) []int {
    bits := make([]int, 0)

    for n > 0 {
        bit := n & 1 // Extract the least significant bit
        bits = append(bits, bit)
        n >>= 1      // Shift right by 1 bit
    }

    // Reverse the bit slice to get the correct order
    for i, j := 0, len(bits)-1; i < j; i, j = i+1, j-1 {
        bits[i], bits[j] = bits[j], bits[i]
    }

    return bits
}

