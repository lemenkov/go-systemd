package main

import (
	"fmt"
	"bufio"
	"encoding/binary"
	//"reflect"
	//"bytes"
	//"errors"
	"unsafe"
)

const (
	OBJECT_UNUSED = iota
	OBJECT_DATA = iota
	OBJECT_FIELD = iota
	OBJECT_ENTRY = iota
	OBJECT_DATA_HASH_TABLE = iota
	OBJECT_FIELD_HASH_TABLE = iota
	OBJECT_ENTRY_ARRAY = iota
	OBJECT_TAG = iota
)

const (
	OBJECT_COMPRESSED_XZ = 1 << iota
	OBJECT_COMPRESSED_LZ4 = 1 << iota
)

type ObjectPayload interface {
	// r - A connected reader
	// payload_size - a size of an ObjectPayload including all its fields
	Read(r *bufio.Reader, payload_size le64_t) error

	// FIXME
	//Write(r *bufio.Reader)

	PrettyPrint()
}

type ObjectHeader struct {
	Type uint8_t
	Flags uint8_t
	Reserved [6]uint8_t
	Size le64_t // The entire size of an object (including all the headers)
}

type UnusedObject struct {
	// Just a padding up to 64 bytes (16 bytes of a header + these 48 extra bytes)
	Payload [48]uint8_t
}

func (obj *UnusedObject) Read(r *bufio.Reader, payload_size le64_t) error {
	return binary_read(r, binary.LittleEndian, obj)
}

func (obj UnusedObject) PrettyPrint() {
	fmt.Printf("Type: OBJECT_UNUSED\n")
}

type DataObject struct {
	hash le64_t
	next_hash_offset le64_t
	next_field_offset le64_t
	entry_offset le64_t
	entry_array_offset le64_t
	n_entries le64_t
	payload []uint8_t
	padding []uint8_t
}

func (obj *DataObject) Read(r *bufio.Reader, payload_size le64_t) error {
	obj.payload = make([]uint8_t, payload_size - 6*8)
	obj.padding = make([]uint8_t, count_padding(payload_size - 6*8))
	for _, elem := range []interface{}{
		&obj.hash,
		&obj.next_hash_offset,
		&obj.next_field_offset,
		&obj.entry_offset,
		&obj.entry_array_offset,
		&obj.n_entries,
		&obj.payload,
		&obj.padding,
	} {
		func (data interface{}) {
			binary_read(r, binary.LittleEndian, data)
		} (elem)
	}

	// FIXME
	return nil
}

func (obj DataObject) PrettyPrint() {
	fmt.Printf("Type: OBJECT_DATA, Payload: [%s] %#v\n", obj.payload, obj.payload)
}

type FieldObject struct {
	hash le64_t
	next_hash_offset le64_t
	head_data_offset le64_t
	payload []uint8_t
	padding []uint8_t
}

func (obj *FieldObject) Read(r *bufio.Reader, payload_size le64_t) error {
	obj.payload = make([]uint8_t, payload_size - 3*8)
	obj.padding = make([]uint8_t, count_padding(payload_size - 3*8))
	for _, elem := range []interface{}{
		&obj.hash,
		&obj.next_hash_offset,
		&obj.head_data_offset,
		&obj.payload,
		&obj.padding,
	} {
		func (data interface{}) {
			binary_read(r, binary.LittleEndian, data)
		} (elem)
	}

	// FIXME
	return nil
}

func (obj FieldObject) PrettyPrint() {
	fmt.Printf("Type: OBJECT_FIELD, Payload: [%s] %#v\n", obj.payload, obj.payload)
}

type EntryItem struct {
	ObjectOffset le64_t
	Hash le64_t
}

type EntryObject struct {
	seqnum le64_t
	realtime le64_t
	monotonic le64_t
	boot_id sd_id128
	xor_hash le64_t
	items []EntryItem
}

func (obj *EntryObject) Read(r *bufio.Reader, payload_size le64_t) error {
	obj.items = make([]EntryItem, (payload_size - 6*8) / 16)
	for _, elem := range []interface{}{
		&obj.seqnum,
		&obj.realtime,
		&obj.monotonic,
		&obj.boot_id,
		&obj.xor_hash,
	} {
		func (data interface{}) {
			binary_read(r, binary.LittleEndian, data)
		} (elem)
	}
	for i := 0; i < len(obj.items); i++ {
		binary_read(r, binary.LittleEndian, &obj.items[i])
	}

	// FIXME
	return nil
}

func (obj EntryObject) PrettyPrint() {
	fmt.Printf("Type: OBJECT_ENTRY\n")
	//fmt.Printf("EntryItem: %d %d\n", obj.object_offset, obj.hash)
}


type HashItem struct {
	HeadHashOffset le64_t
	TailHashOffset le64_t
}

type HashTableObject struct {
	items []HashItem
}

func (obj *HashTableObject) Read(r *bufio.Reader, payload_size le64_t) error {
	obj.items = make([]HashItem, payload_size / 16)
	for i := 0; i < len(obj.items); i++ {
		binary_read(r, binary.LittleEndian, &obj.items[i])
	}

	// FIXME
	return nil
}

func (obj HashTableObject) PrettyPrint() {
	fmt.Printf("Type: OBJECT_DATA_HASH_TABLE / OBJECT_FIELD_HASH_TABLE\n")
	//fmt.Printf("Read item # %d\n", i)
}

type EntryArrayObject struct {
	NextEntryArrayOffset le64_t
	Items []le64_t
}

func (obj *EntryArrayObject) Read(r *bufio.Reader, payload_size le64_t) error {
	obj.Items = make([]le64_t, (payload_size - 8) / 8)
	binary_read(r, binary.LittleEndian, &obj.NextEntryArrayOffset)
	for i := 0; i < len(obj.Items); i++ {
		binary_read(r, binary.LittleEndian, &obj.Items[i])
	}

	// FIXME
	return nil
}

func (obj EntryArrayObject) PrettyPrint() {
	fmt.Printf("Type: OBJECT_ENTRY_ARRAY\n")
	//fmt.Printf("EntryArrayObject: %d %d\n", payload_size, payload_size - 8)
	//fmt.Printf("EntryArrayItem: %d\n", obj.items[i])
}

const TAG_LENGTH = (256/8) // SHA-256 HMAC

type TagObject struct {
	Seqnum le64_t
	Epoch le64_t
	Tag [TAG_LENGTH]uint8_t
}

func (obj *TagObject) Read(r *bufio.Reader, payload_size le64_t) error {
	return binary_read(r, binary.LittleEndian, obj)
}

func (obj TagObject) PrettyPrint() {
	fmt.Printf("Type: OBJECT_TAG, Payload: [%s], %#v\n", obj.Tag, obj.Tag)
}

type Object struct {
	header ObjectHeader
	payload ObjectPayload
}

func read_journal_objects(r *bufio.Reader, header JournalHeader) []Object {
	var objects []*Object
	var o *Object
	var e error

	fmt.Printf("E, %#v\n", e)
	o, e = read_journal_object(r)
	for e == nil {
		objects = append(objects, o)
		o, e = read_journal_object(r)
	}
	fmt.Printf("E1, %#v\n", e)
	fmt.Printf("O, %#v\n", len(objects))
	return nil
}

func read_journal_object(r *bufio.Reader) (*Object, error) {
	var obj Object
	var e error

	// Read the header first
	e = binary_read(r, binary.LittleEndian, &obj.header)
	if e == nil {
		switch obj.header.Type {
			case OBJECT_UNUSED:
				obj.payload = &UnusedObject{}
			case OBJECT_DATA:
				obj.payload = &DataObject{}
			case OBJECT_FIELD:
				obj.payload = &FieldObject{}
			case OBJECT_ENTRY:
				obj.payload = &EntryObject{}
			case OBJECT_DATA_HASH_TABLE:
				obj.payload = &HashTableObject{}
			case OBJECT_FIELD_HASH_TABLE:
				obj.payload = &HashTableObject{}
			case OBJECT_ENTRY_ARRAY:
				obj.payload = &EntryArrayObject{}
			case OBJECT_TAG:
				obj.payload = &TagObject{}
			default:
				fmt.Printf("Type: %X\n", obj.header.Type)
		}

		e = obj.payload.Read(r, obj.header.Size - le64_t(unsafe.Sizeof(obj.header)))
		if e == nil && obj.header.Type != OBJECT_UNUSED {
			//obj.payload.PrettyPrint()
			return &obj, nil
		}
	}
	return nil, e
}

