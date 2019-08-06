package dbfparse

import (
	"os"
	"reflect"
	"testing"
)

func TestDbParse(t *testing.T) {
	fileName := "FSDHZ.DBF"
	fp, err := os.Open(fileName)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer fp.Close()

	parser, err := NewParser(fp)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Logf("The length of header is %v", parser)
		for _, fieldDesc := range parser.FieldDescs {
			t.Logf("field desc %v", fieldDesc)
		}

	}

	type Test struct {
		A string `field:"BH"`
		B string `field:"LH"`
		C string `field:"NAMEA"`
		D string `field:"DW"`
		E string `field:"FF"`
		F string `field:"DF"`
		G string `field:"BYSZ"`
		H string `field:"SYSZ"`
		I string `field:"SF"`
		J string `field:"QF"`
		K string `field:"HJ"`
		L string `field:"QIF"`
		M string `field:"ZA"`
		N string `field:"ZH"`
	}

	Register(&Test{})

	recordChan, err := parser.ParseRecord("Test")
	if err != nil {
		t.Error(err.Error())
	} else {
		i := 0
		for recordI := range recordChan {
			if i > 10 {
				continue
			}
			record, ok := recordI.(*Test)
			if ok {
				t.Logf("%v", record)
			}
			i++
		}
		t.Logf("ok")
	}
}

func TestNewObject(t *testing.T) {
	type User struct {
		Name string `tag:"name"`
		Age  int    `tag:"age"`
	}
	Register(&User{})
	ui, err := NewObject("User")
	if err != nil {
		t.Error(err.Error())
	}

	tpy := reflect.TypeOf(ui)
	v := reflect.ValueOf(ui)

	v.Elem().Field(0).SetString("aaaa")
	v.Elem().Field(1).SetInt(10)

	t.Logf("%T, %v, %v", ui, ui, tpy.Elem().Field(0).Tag.Get("tag"))

}
