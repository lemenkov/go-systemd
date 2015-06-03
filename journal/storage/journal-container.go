package main

import (
	"fmt"
	//"encoding/binary"
	"bufio"
	//"bytes"
	//"errors"
)

// https://github.com/systemd/systemd/blob/ff9b60f/src/journal/journal-def.h#L191

func read_journal(r *bufio.Reader) {
	// Check if it's a systemd journal file first
	is_journal(r)

	// Ok this is likely a systemd journal file
	// Now let's read and parse header starting from signature
	header := read_journal_header(r)

	//
	// Print information about Journal header
	//
	//fmt.Printf("%#v\n", header)
	fmt.Printf("Objects: %d, Entries: %d, Fields: %d, Tags: %d\n", header.n_objects, header.n_entries, header.n_fields, header.n_tags)

	// Read consequent objects
	read_journal_objects(r, header)
}
