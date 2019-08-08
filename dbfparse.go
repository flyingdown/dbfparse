package dbfparse

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/flyingdown/codeconvert"
)

type char uint8

type DbfHeader struct {
	Version         byte
	LastYear        uint16
	LastMonth       uint8
	LastDay         uint8
	NumberOfRec     uint32
	HeaderLength    uint16
	RecordLength    uint16
	TransactionFalg byte
	EncryptionFlag  byte
	FreeRecThread   uint32
	MDXFlag         byte
	LangDriver      byte
}

type FieldDesc struct {
	FieldName      string
	FieldType      char
	FieldLength    uint8
	FieldPrecision uint8
}

type parser struct {
	ReadSeeker io.ReadSeeker
	DbfHeader
	FieldDescs []*FieldDesc
}

func NewParser(readSeeker io.ReadSeeker) (*parser, error) {
	parser := &parser{
		ReadSeeker: readSeeker,
		FieldDescs: []*FieldDesc{},
	}

	err := parser.ParseHead()
	if err != nil {
		return nil, err
	}

	return parser, nil
}

func (p *parser) ParseHead() error {
	fp := p.ReadSeeker
	buf := make([]byte, 2)
	fp.Seek(8, 0) // nolint: errcheck
	_, err := fp.Read(buf)
	if err != nil {
		return err
	}

	headerLength := uint16(buf[1])<<8 | uint16(buf[0])

	p.HeaderLength = headerLength

	// fmt.Println("The length of header is", p.HeaderLength)
	fp.Seek(0, 0) // nolint: errcheck
	buf = make([]byte, p.HeaderLength)
	_, err = fp.Read(buf)
	if err != nil {
		return err
	}

	p.Version = buf[0]

	p.LastYear = uint16(buf[1]) + 0x076c
	p.LastMonth = buf[2]
	p.LastDay = buf[3]

	p.NumberOfRec = uint32(buf[7])<<24 | uint32(buf[6])<<16 | uint32(buf[5])<<8 | uint32(buf[4])

	p.HeaderLength = uint16(buf[9])<<8 | uint16(buf[8])

	p.RecordLength = uint16(buf[11])<<8 | uint16(buf[10])

	p.TransactionFalg = buf[14]

	p.EncryptionFlag = buf[15]

	p.FreeRecThread = uint32(buf[19])<<24 | uint32(buf[18])<<16 | uint32(buf[17])<<8 | uint32(buf[16])

	p.MDXFlag = buf[28]

	p.LangDriver = buf[29]

	for curLen := 32; buf[curLen] != 0x0D; curLen += 32 {
		nameBuf, _ := codeconvert.GbkToUtf8(buf[curLen : curLen+11])
		fieldDesc := &FieldDesc{
			FieldName:      strings.Trim(string(nameBuf), "\x00"),
			FieldType:      char(buf[curLen+11]),
			FieldLength:    buf[curLen+16],
			FieldPrecision: buf[curLen+17],
		}
		p.FieldDescs = append(p.FieldDescs, fieldDesc)
	}

	return nil

}

func (p *parser) ParseRecord(name string) (chan interface{}, error) {
	fp := p.ReadSeeker
	fp.Seek(int64(p.HeaderLength), 0) // nolint: errcheck

	// init the struct
	tmp, err := NewObject(name)
	if err != nil {
		return nil, err
	}

	// find Field tag and set map
	type structInfo struct {
		index int
		kind  reflect.Kind
	}
	t := reflect.TypeOf(tmp)
	if t.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s'type is not struct, %s", name, t)
	}
	infoMap := map[string]structInfo{}
	for i := 0; i < t.Elem().NumField(); i++ {
		field := t.Elem().Field(i)
		tag := field.Tag.Get("field")
		for _, fieldDesc := range p.FieldDescs {
			if tag == fieldDesc.FieldName {
				infoMap[tag] = structInfo{
					i,
					field.Type.Kind(),
				}
			}
		}
	}

	recordChan := make(chan interface{})

	go func() {
		defer close(recordChan)
		for i := 0; i < int(p.NumberOfRec); i++ {
			// for i := 0; i < 1; i++ {
			// init the struct
			record, err := NewObject(name)
			if err != nil {
				log.Println(err)
				continue
			}
			recordValue := reflect.ValueOf(record).Elem()

			buf := make([]byte, p.RecordLength)
			n, err := fp.Read(buf)
			if n != int(p.RecordLength) || err != nil {
				log.Printf("parse error, read %d, %s", n, err.Error())
				continue
			}

			// This record is deleted
			if buf[0] == 0x2a {
				continue
			}

			// jump over the delete flag
			var curLen uint8 = 1
			for _, fieldDesc := range p.FieldDescs {
				begin := curLen
				curLen += fieldDesc.FieldLength
				switch fieldDesc.FieldType {
				case 'C':
					valueBuf, _ := codeconvert.GbkToUtf8(buf[begin:curLen])
					// fmt.Printf("[name: %s, value: %s]", fieldDesc.FieldName, strings.TrimSpace(string(valueBuf)))
					info, ok := infoMap[fieldDesc.FieldName]
					if ok || info.kind == reflect.String {
						recordValue.Field(info.index).SetString(strings.TrimSpace(string(valueBuf)))
					}
				case 'N':
					valueBuf, _ := codeconvert.GbkToUtf8(buf[begin:curLen])
					// fmt.Printf("[name: %s, value: %s]", fieldDesc.FieldName, strings.TrimSpace(string(valueBuf)))
					info, ok := infoMap[fieldDesc.FieldName]
					if ok || info.kind == reflect.String {
						recordValue.Field(info.index).SetString(strings.TrimSpace(string(valueBuf)))
					}
					// case 'I':
					// 	fmt.Printf("name %s, value %f\n", fieldDesc.FieldName, buf[start:curLen])

				}

			}
			// fmt.Println(record)
			recordChan <- record
		}
	}()
	return recordChan, nil
}
