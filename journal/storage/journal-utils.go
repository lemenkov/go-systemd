package main

import (
	//"flag"
	"fmt"
	"io"
	//"io/ioutil"
	//"bufio"
	"encoding/binary"
	//"os"
	//"reflect"
	//"time"
)

func check(e error) {
	if e != nil {
		fmt.Printf("ERROR: %s\n", e)
		//panic(e)
	}
}

func binary_read(r io.Reader, order binary.ByteOrder, data interface{}) error {
	err := binary.Read(r, order, data)
	//check(err)
	return err
}

func count_padding(size le64_t) int {
	rest := size % 8
	if rest == 0 {
		return 0
	} else {
		return int(8 - rest)
	}
}
