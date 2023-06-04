
package main

import (
    "fmt"
    "os"
    "strconv"
    "sync"
    "math"
    "time"
    "encoding/hex"
    "runtime"
)

type PayloadSet map[string]struct{}

type PayloadGroup struct {

	mutex sync.Mutex
	payload_set PayloadSet
	candidate_counter int

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

func produce_payloads_int_range ( ints_chan <-chan int , payload_slice_in []string , send_channel chan<- string , wg *sync.WaitGroup ) {

	fmt.Println( "starting job at " , start_int )
	
	defer fmt.Println( "finished job at " , end_int )
	defer wg.Done()
	
	for {
		combo_num , channel_open := <-ints_chan
		if !channel_open {
			return
		}
		send_channel <- generate_payload_by_binary_int_map( combo_num , payload_slice_in )
		
	}

}

func consume_payloads ( payload_group *PayloadGroup , total_candidates int , receive_channel <-chan string , doneChannel <-chan struct{}, wg *sync.WaitGroup ) {
	
	defer fmt.Println("consumer done")
	defer wg.Done()
	
	for {
		select {
		case payload := <-receive_channel:

				if _ , exists := payload_set[payload]; !exists {
					payload_group.payload_set[payload] = struct{}{}
					fmt.Println("number of payloads: " , len(payload_group.payload_set) )
				}
				payload_group.candidate_counter += 1

		case <-doneChannel:
			return
		}
	}
	

}

const (
	NumProducers  = 5
	NumConsumers  = 3
	MaxGoroutines = 10
	MaxBufferSize = 1000
)

func generate_all_payloads( scheme string , expected_host string , directory string , evil_host string , evil_dir string , evil_port string , fake_password string ) *PayloadGroup {

	var wgProducers sync.WaitGroup
	var wgConsumers sync.WaitGroup
	
	payload_group := PayloadGroup{}
	payload_slice := create_payload_slice( scheme , expected_host , directory , evil_host , evil_dir , evil_port , fake_password )
	//Pow returns float64, but we want integer division at delta definition
	total_payloads := int ( math.Pow( 2 , len( payload_slice ) ) )
	max_goroutines := runtime.NumCPU()

	data_channel := make(chan string, MaxBufferSize)
	done_channel := make(chan struct{})
	ints_channel := make(chan int , total_payloads)

	delta := total_payloads / max_goroutines
	
	for i := 1; i <= total_candidates; i+=delta {
		wgConsumers.Add(1)
		go consume_payloads( &payload_group , total_payloads , ch , &wgConsumers )
	}
	
	for i := 1; i <= total_candidates; i+=delta {
		wgProducers.Add(1)
		go produce_payloads_int_range ( i , i+delta , payload_slice , ch , &wgProducers ) 
		fmt.Println("wg consumer add")
	}
	
	// Monitor data rates and adjust the buffer size
	go monitorDataRates(dataChan, doneChan, &wgProducers, &wgConsumers)
	
	wgConsumers.Wait()
	wgProducers.Wait()
	return &payload_group

}

func generate_payload_by_binary_int_map ( int_in int , payload_slice []string ) ( payload_out string ) {

	index := 0
	for int_in > 0 {
		if int_in & 1 {
			payload_out = fmt.Sprintf( "%s%s" , payload_out , payload_slice[index] )
		}
		int_in >>= 1
		index += 1
	}
	return

}

func monitorDataRates(dataChan <-chan string, doneChan <-chan struct{}, wgProducers *sync.WaitGroup, wgConsumers *sync.WaitGroup) {
	
	currentDataChan := make(chan string, initialBufferSize)
	
	knob_A := 1
	knob_B := 1
	knob_C := 1
	knob_D := 1
	lastProducerLen := 0
	lastConsumerDif := 0

	waiter_counter := 0
	
	second_tick := time.NewTicker(time.Second)
	centisecond_tick := time.NewTicker(10*time.Millisecond)

	for {
		select {
		case <-centisecond_tick.C:
			waiter_counter += wgConsumers.Waiting()
		case <-second_tick.C:
			avg_waiters = float64(waiter_counter)/100.0
			waiter_counter = 0
			if avg_waiters > 1 {
				go produce_payloads_int_range ( i , i+delta , payload_slice , ch , &wgProducers ) 
				fmt.Println("wg consumer add")
			}
			// Calculate the current data rates
//			producerDataLen := float64(len(dataChan)) 
//			activeConsumerDif := float64(NumConsumers) - float64(wgConsumers.Waiting())
//
//			// Calculate the desired buffer size based on the data rates
//			desiredBufferSize := int(producerDataRate + consumerDataRate)
//			desiredGoroutines := int(producerDataRate + consumerDataRate)
//
//			// Limit the desired buffer size to the maximum allowed
//			if desiredBufferSize > MaxBufferSize {
//				desiredBufferSize = MaxBufferSize
//			}
//
//			// Check if the buffer size needs to be increased
//			if desiredBufferSize > cap(currentDataChan) {
//				// Create a new channel with the desired buffer size
//				newDataChan := make(chan string, desiredBufferSize)
//
//				// Start a goroutine to drain data from the current channel and send it to the new channel
//				go func() {
//					for data := range currentDataChan {
//						newDataChan <- data
//					}
//					close(newDataChan)
//				}()
//
//				// Replace the current channel with the new channel
//				currentDataChan = newDataChan
//			}
//
//			// Adjust the buffer size of the original data channel
//			dataChan = currentDataChan

		case <-doneChan:
			return
		}
	}
}


func create_payload_slice ( scheme string , expected_host string , directory string , evil_host string , evil_dir string , evil_port string , fake_password string ) ( payload_slice []string ) {

	payload_slice = []string{ scheme,"/","\\","/","\\","/","\\","/","\\",evil_host,evil_dir,".","#",".",expected_host,"@",directory, fmt.Sprintf(".%s",evil_host) , fmt.Sprintf(":%s",fake_password),"@", fmt.Sprintf("@%s",evil_host) , evil_dir }
	return

}
