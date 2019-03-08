package fixcsv

import (
	"bufio"
	"bytes"
	"encoding"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
)

var (
	delimeter = "||"
)

func Marshal(v interface{}) ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	err := NewEncoder(buff).Encode(v)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

// MarshalInvalidTypeError describes an invalid type being marshaled.
type MarshalInvalidTypeError struct {
	typeName string
}

func (e *MarshalInvalidTypeError) Error() string {
	return "fixcsv: cannot marshal unknown Type " + e.typeName
}

type Encoder struct {
	w *bufio.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		bufio.NewWriter(w),
	}
}

func (e *Encoder) Encode(i interface{}) (err error) {
	if i == nil {
		return nil
	}

	v := reflect.ValueOf(i)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice {
		err = e.writeLines(v)
	} else {
		err = e.writeLine(reflect.ValueOf(i))
	}
	if err != nil {
		return err
	}
	return e.w.Flush()
}

func (e *Encoder) writeLines(v reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		err := e.writeLine(v.Index(i))
		if err != nil {
			return err
		}

		if i != v.Len()-1 {
			_, err := e.w.Write([]byte("\n"))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *Encoder) writeLine(v reflect.Value) (err error) {
	b, err := newValueEncoder(v.Type())(v)
	if err != nil {
		return err
	}
	_, err = e.w.Write(b)
	return err
}

type valueEncoder func(v reflect.Value) ([]byte, error)

func newValueEncoder(t reflect.Type) valueEncoder {
	if t == nil {
		return nilEncoder
	}
	if t.Implements(reflect.TypeOf(new(encoding.TextMarshaler)).Elem()) {
		return textMarshalerEncoder
	}

	switch t.Kind() {
	case reflect.Ptr, reflect.Interface:
		return ptrInterfaceEncoder
	case reflect.Struct:
		return structEncoder
	case reflect.String:
		return stringEncoder
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return intEncoder
	case reflect.Float64:
		return floatEncoder(0, 64)
	case reflect.Float32:
		return floatEncoder(0, 32)
	}
	return unknownTypeEncoder(t)
}

func structEncoder(v reflect.Value) ([]byte, error) {
	var specs []fieldSpec
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		var (
			err      error
			spec     fieldSpec
			errParse error
		)
		spec.position, spec.length, errParse = parseTag(f.Tag.Get("fixcsv"))
		if errParse != nil {
			continue
		}
		spec.value, err = newValueEncoder(f.Type)(v.Field(i))
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return encodeSpecs(specs)
}

type fieldSpec struct {
	position int
	length   int
	value    []byte
}

type fieldSpecSorter []fieldSpec

func (s fieldSpecSorter) Len() int           { return len(s) }
func (s fieldSpecSorter) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s fieldSpecSorter) Less(i, j int) bool { return s[i].position < s[j].position }

func encodeSpecs(specs []fieldSpec) (byt []byte, err error) {
	sort.Sort(fieldSpecSorter(specs))

	data := bytes.NewBuffer([]byte{})
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error encode %#v", r)
		}
	}()
	for idx, spec := range specs {
		if len(spec.value) > spec.length {
			data.Write(spec.value[:spec.length])
		} else {
			data.Write(spec.value)
		}
		if idx != len(specs)-1 {
			data.WriteString(delimeter)
		}
	}
	byt = data.Bytes()
	return
}

func textMarshalerEncoder(v reflect.Value) ([]byte, error) {
	return v.Interface().(encoding.TextMarshaler).MarshalText()
}

func ptrInterfaceEncoder(v reflect.Value) ([]byte, error) {
	if v.IsNil() {
		return nilEncoder(v)
	}
	return newValueEncoder(v.Elem().Type())(v.Elem())
}

func stringEncoder(v reflect.Value) ([]byte, error) {
	return []byte(v.String()), nil
}

func intEncoder(v reflect.Value) ([]byte, error) {
	return []byte(strconv.Itoa(int(v.Int()))), nil
}

func floatEncoder(perc, bitSize int) valueEncoder {
	return func(v reflect.Value) ([]byte, error) {
		return []byte(strconv.FormatFloat(v.Float(), 'f', perc, bitSize)), nil
	}
}

func nilEncoder(v reflect.Value) ([]byte, error) {
	return nil, nil
}

func unknownTypeEncoder(t reflect.Type) valueEncoder {
	return func(value reflect.Value) ([]byte, error) {
		return nil, &MarshalInvalidTypeError{typeName: t.Name()}
	}
}
