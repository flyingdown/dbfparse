package dbfparse

import (
	"testing"
)

func TestDbParse(t *testing.T) {
	fileName := "FSDHZ.DBF"
	parser, err := NewParser(fileName)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Logf("The length of header is %v", parser)
		for _, fieldDesc := range parser.FieldDescs {
			t.Logf("field desc %v", fieldDesc)
		}

	}
	err = parser.ParseRecord()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Logf("ok")
	}
}
