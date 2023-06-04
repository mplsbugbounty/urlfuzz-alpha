
package main

import (
    "fmt"
    "gonum.org/v1/gonum/stat/combin"
    "os"
    "strconv"
    "sync"
    "math"
    "time"
    "encoding/hex"
)

type PayloadGroup struct {

	mutex sync.Mutex
	payloads []string

}

func convert_string_to_complete_percent_encoded_hex( str_in string ) ( encoded_str_out string ) {
	
	encoded_str_out = ""
	for i := 0; i < len( str_in ); i++ { 
		encoded_str_out = fmt.Sprintf( "%s%s%s", encoded_str_out , "%" , hex.EncodeToString( []byte{str_in[i]} ) )
	}
	return
}

func main() {

	scheme := os.Args[1]
	host := os.Args[2]
	port := os.Args[3]
	directory := os.Args[4]
	fragment := os.Args[5]
	split_url := []string{ scheme , host , port , directory , fragment }
	split_url = split_url
	start := time.Now()
	evil_elements := []string{ "127.0.0.1:8080" , "." , "/","\\" }
	payload_elements := create_payload_elements( evil_elements )
	all_permutations := generate_all_permutations( payload_elements )
	fmt.Println( all_permutations )	
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println("All permutations generated in " , elapsed )

}

func create_payload_elements ( evil_elements []string ) ( payload_elements_out []string ) {
	
	payload_elements_out = append( payload_elements_out , evil_elements... )
	working_slice := []string{}
	for _ , element := range( evil_elements ) {
		working_slice = append( working_slice , convert_string_to_complete_percent_encoded_hex( element ) )
	}
	payload_elements_out = append( payload_elements_out , working_slice... )
	return
}

func powerset_permutations_cardinality ( n int ) ( cardinality int ) {
	
	//Gamma(n+1) = n! for integer n
	n_factorial := math.Gamma( float64( n + 1 ) )
	//The set of all permutations of elements of the powerset of X is given by C = Sum(n!/k!)
	running_sum := 0.0
	for k := 0; k <= n ; k++ {
		running_sum = running_sum + n_factorial/math.Gamma( float64( k + 1 ) )
	}
	//subtract 1 because we aren't including the empty set in this context.
	cardinality = int(running_sum) - 1
	return

}

func produce_string_n_element_permutations ( str_slice_in []string , n int , send_channel chan<- string , wg *sync.WaitGroup ) {

	numstrings := len(str_slice_in)
	fmt.Println("starting job " + strconv.Itoa(n) + " of " + strconv.Itoa(numstrings))
	gen := combin.NewPermutationGenerator( numstrings , n )
	idx := 0
	for gen.Next() {
		this_permutation := gen.Permutation(nil)
		idx++
		working_string := ""
		for _ , inner_ind := range(this_permutation) {
			working_string = fmt.Sprintf( "%s%s" , working_string , str_slice_in[inner_ind])
		}
		send_channel <- working_string
		fmt.Println("working_string:" + working_string)
	}
	fmt.Println("finished job " + strconv.Itoa(n) + " of " + strconv.Itoa(numstrings))
	wg.Done()

}

func consume_string_permutations ( payload_group *PayloadGroup , cardinality int , receive_channel <-chan string , wg *sync.WaitGroup ) {
	
	for permutation := range receive_channel {
		fmt.Println("number of payloads: " , len(payload_group.payloads) )
		payload_group.payloads = append(payload_group.payloads  , permutation)
		fmt.Println("finished permutation " + strconv.Itoa(len(payload_group.payloads)) + " of " + strconv.Itoa(cardinality))
		if len(payload_group.payloads ) == cardinality {
			fmt.Println("break triggered")
			break
		}
	}
	fmt.Println("consumer done")
	wg.Done()

}

func generate_all_permutations( str_slice_in []string ) *PayloadGroup {

	var wg sync.WaitGroup
	numstrings := len(str_slice_in)
	payload_group := PayloadGroup{}
	
	ch := make(chan string)
	defer close(ch)
	
	//fmt.Println("consumer wg add")
	wg.Add(1)
	go consume_string_permutations( &payload_group , powerset_permutations_cardinality( numstrings ) , ch , &wg )

	for i := 1; i <= numstrings; i++ {
		wg.Add(1)
		fmt.Println("wg add")
		go produce_string_n_element_permutations( str_slice_in , i , ch , &wg ) 
	}
	wg.Wait()
	return &payload_group

}
