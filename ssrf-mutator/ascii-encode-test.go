package main

import (
	"os"
	"fmt"
	"encoding/hex"
)

func convert_string_to_complete_percent_encoded_hex( str_in string ) ( encoded_str_out string ) {
    
        encoded_str_out = ""
        for i := 0; i < len( str_in ); i++ { 
		encoded_str_out = fmt.Sprintf( "%s%s%s", encoded_str_out , "%" , hex.EncodeToString( []byte{str_in[i]} ) )  
        } 
	return
}

func main() {

	fmt.Println(convert_string_to_complete_percent_encoded_hex(os.Args[1]))

}
