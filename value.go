package json

import (
	"errors"
)

type ValueType int

const (
	ValueNull ValueType = iota
	ValueFalse
	ValueTrue
	ValueNumber
	ValueString
	ValueArray
	ValueObject
)

// jsonValueError 访问 jsonValue 时的error
type jsonValueError struct {
	msg string // description of error
}

func (e *jsonValueError) Error() string { return e.msg }

// jsonValue 解析json的中转值，[]byte > jsonValue -> var
type jsonValue struct {
	object
	array
	s         []byte
	n         float64
	valueType ValueType
}

type object struct {
	size   int
	keys   []*jsonValue
	values []*jsonValue
}

type array struct {
	len    int
	values []*jsonValue
}

func (v *jsonValue) getValueType() ValueType {
	return v.valueType
}

func (v *jsonValue) getBoolean() (bool, error) {
	if v.valueType != ValueTrue && v.valueType != ValueFalse {
		return false, v.error("value type isn't boolean")
	}
	return v.valueType == ValueTrue, nil
}

func (v *jsonValue) getNumber() (float64, error) {
	if v.valueType != ValueNumber {
		return 0.0, v.error("value type isn't number")
	}
	return v.n, nil
}

func (v *jsonValue) getString() (string, error) {
	if v.valueType != ValueString {
		return "", v.error("value type isn't string")
	}
	return string(v.s), nil
}

func (v *jsonValue) getArrayLen() int {
	return v.array.len
}

func (v *jsonValue) getArrayElem(index int) (*jsonValue, error) {
	if v.valueType != ValueArray {
		return nil, v.error("value type isn't array")
	}
	if index > v.array.len-1 {
		return nil, errors.New("array out range")
	}
	return v.array.values[index], nil
}

func (v *jsonValue) getObjectSize() int {
	return v.object.size
}

func (v *jsonValue) getObjectKey(index int) (*jsonValue, error) {
	if v.valueType != ValueObject {
		return nil, v.error("value type isn't object")
	}
	if index > v.object.size-1 {
		return nil, v.error("object out range")
	}
	return v.object.keys[index], nil
}

func (v *jsonValue) getObjectValue(index int) (*jsonValue, error) {
	if v.valueType != ValueObject {
		return nil, v.error("value type isn't object")
	}
	if index > v.object.size-1 {
		return nil, v.error("object out range")
	}
	return v.object.values[index], nil
}

func (v *jsonValue) error(msg string) error {
	return &jsonValueError{msg}
}
