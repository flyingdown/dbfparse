package dbfparse

import (
	"os"
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
	FileName string
	Fp       *os.File
	DbfHeader
	FieldDescs []*FieldDesc
}

func NewParser(fileName string) (*parser, error) {
	fp, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	parser := &parser{
		FileName:   fileName,
		Fp:         fp,
		FieldDescs: []*FieldDesc{},
	}

	err = parser.ParseHead()
	if err != nil {
		return nil, err
	}

	return parser, nil
}

func (p *parser) ParseHead() error {
	fp := p.Fp
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
		fieldDesc := &FieldDesc{
			FieldName:      string(buf[curLen : curLen+11]),
			FieldType:      char(buf[curLen+11]),
			FieldLength:    buf[curLen+16],
			FieldPrecision: buf[curLen+17],
		}
		p.FieldDescs = append(p.FieldDescs, fieldDesc)
	}

	return nil

}

func (p *parser) ParseRecord() error {
	fp, err := os.Open(p.FileName)
	if err != nil {
		return err
	}

	defer fp.Close()
	fp.Seek(int64(p.HeaderLength), 0) // nolint: errcheck

	// record, err := NewObject(name)
	// if err != nil {
	// 	return err
	// }

	// for i := 0; i < int(p.NumberOfRec); i++ {
	for i := 0; i < 1; i++ {
		buf := make([]byte, p.RecordLength)
		n, err := fp.Read(buf)
		if n != int(p.RecordLength) || err != nil {
			return fmt.Errorf("parse error, read %d, %s", n, err.Error())
		}

		// This record is deleted
		if buf[0] == 0x2a {
			continue
		}

		var (
			start  uint8 = 0
			curLen uint8 = 0
		)
		for _, fieldDesc := range p.FieldDescs {
			start = curLen
			curLen += fieldDesc.FieldLength
			switch fieldDesc.FieldType {
			case 'C':
				fmt.Printf("[name %s, value %s]", fieldDesc.FieldName, string(buf[start:curLen]))
			case 'N':
				fmt.Printf("[name %s, value %s]", fieldDesc.FieldName, string(buf[start:curLen]))
				// case 'I':
				// 	fmt.Printf("name %s, value %f\n", fieldDesc.FieldName, buf[start:curLen])

			}
		}
		fmt.Println()

	}
	return nil
}
