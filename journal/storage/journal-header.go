package main

import (
	"fmt"
	"encoding/binary"
	"bufio"
	"bytes"
	"errors"
)

// https://github.com/systemd/systemd/blob/ff9b60f/src/journal/journal-def.h#L191

const (
	STATE_OFFLINE = iota
	STATE_ONLINE = iota
	STATE_ARCHIVED = iota
)

const (
	HEADER_INCOMPATIBLE_COMPRESSED_XZ = 1 << iota
	HEADER_INCOMPATIBLE_COMPRESSED_LZ4 = 1 << iota
)

const (
	HEADER_COMPATIBLE_SEALED = iota
)

const HEADER_SIGNATURE =  "LPKSHHRH"

type JournalHeader struct {
	signature [8]uint8_t
	compatible_flags le32_t
	incompatible_flags le32_t
	state uint8_t
	reserved [7]uint8_t
	file_id sd_id128
	machine_id sd_id128
	boot_id sd_id128
	seqnum_id sd_id128
	header_size le64_t
	arena_size le64_t
	data_hash_table_offset le64_t
	data_hash_table_size le64_t
	field_hash_table_offset le64_t
	field_hash_table_size le64_t
	tail_object_offset le64_t
	n_objects le64_t
	n_entries le64_t
	tail_entry_seqnum le64_t
	head_entry_seqnum le64_t
	entry_array_offset le64_t
	head_entry_realtime le64_t
	tail_entry_realtime le64_t
	tail_entry_monotonic le64_t

	// Added in 187
	n_data le64_t
	n_fields le64_t

	// Added in 189
	n_tags le64_t
	n_entry_arrays le64_t
}

const FSS_HEADER_SIGNATURE = "KSHHRHLP"

// Check if reader is connected to a systemd journal file
func is_journal(r *bufio.Reader) {
	// https://github.com/systemd/systemd/blob/ff9b60f/src/journal/journal-def.h#L189
	PossibleMagic, err := r.Peek(len(HEADER_SIGNATURE))
	check(err)
	if bytes.Equal(PossibleMagic, []byte(HEADER_SIGNATURE)) {
		fmt.Printf("Found header: %s\n", HEADER_SIGNATURE)
	} else {
		panic(errors.New(fmt.Sprintf("Unknown header: %v\n", PossibleMagic)))
	}
}

// Reads and parses header starting from signature
func read_journal_header(r *bufio.Reader) JournalHeader {
	var jh JournalHeader

	for _, elem := range []interface{}{
		&jh.signature,
		&jh.compatible_flags,
		&jh.incompatible_flags,
		&jh.state,
		&jh.reserved,
		&jh.file_id,
		&jh.machine_id,
		&jh.boot_id,
		&jh.seqnum_id,
		&jh.header_size,
		&jh.arena_size,
		&jh.data_hash_table_offset,
		&jh.data_hash_table_size,
		&jh.field_hash_table_offset,
		&jh.field_hash_table_size,
		&jh.tail_object_offset,
		&jh.n_objects,
		&jh.n_entries,
		&jh.tail_entry_seqnum,
		&jh.head_entry_seqnum,
		&jh.entry_array_offset,
		&jh.head_entry_realtime,
		&jh.tail_entry_realtime,
		&jh.tail_entry_monotonic,
	} {
		func (data interface{}) {
			err := binary.Read(r, binary.LittleEndian, data)
			check(err)
		} (elem)
	}

	if jh.header_size > 208 {
		for _, elem := range []interface{}{ &jh.n_data, &jh.n_fields }{
			func (data interface{}) {
				err := binary.Read(r, binary.LittleEndian, data)
				check(err)
			} (elem)
		}
	}

	if jh.header_size > 224 {
		for _, elem := range []interface{}{ &jh.n_tags, &jh.n_entry_arrays }{
			func (data interface{}) {
				err := binary.Read(r, binary.LittleEndian, data)
				check(err)
			} (elem)
		}
	}

	return jh
}
