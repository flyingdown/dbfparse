package dbfparse

import (
	"testing"
)

func TestDbParse(t *testing.T) {
	fileName := "D:/bankofjiaozuo/MAFE平台/工业路支行代扣/FDHZ.DBF"
	parser, err := NewParser(fileName)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Logf("The length of header is %d", parser.HeaderLength)
	}
}
