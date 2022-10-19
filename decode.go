package json

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

// A SyntaxError is a description of a JSON syntax error.
// Unmarshal will return a SyntaxError if the JSON can't be parsed.
type SyntaxError struct {
	msg string // description of error
}

func (e *SyntaxError) Error() string { return e.msg }

// An UnmarshalTypeError describes a JSON jsonValue that was
// not appropriate for a jsonValue of a specific Go type.
type UnmarshalTypeError struct {
	Value string       // description of JSON jsonValue - "bool", "array", "number -5"
	Type  reflect.Type // type of Go jsonValue it could not be assigned to
}

func (e *UnmarshalTypeError) Error() string {
	return "json: cannot unmarshal " + e.Value + " into Go jsonValue of type " + e.Type.String()
}

// 定义解析API
type parse interface {
	// 解析入库
	parser() (*jsonValue, error)
	// 解析value，可以递归调用
	parserValue() (*jsonValue, error)
	// 解析 null、true、false
	parseLiteral(literal []byte, v *jsonValue, valueType ValueType) error
	// 解析数字
	parseNumber(v *jsonValue) error
	// 解析字符串
	parseString(v *jsonValue) error
	// 解析数组
	parseArray(v *jsonValue) error
	// 解析对象
	parseObject(v *jsonValue) error
}

type jsonParse struct {
	data  []byte
	off   int // next read offset in data
	value *jsonValue
}

func (d *jsonParse) init(data []byte) {
	d.data = data
}

func (d *jsonParse) parser() (*jsonValue, error) {
	d.skipWhiteSpace()
	value, err := d.parserValue()
	if err != nil {
		return value, err
	}
	d.skipWhiteSpace()
	if c := d.pop(); c != 0 {
		return value, &SyntaxError{msg: "unexpected end of JSON input"}
	}
	return value, err
}

func (d *jsonParse) parserValue() (*jsonValue, error) {
	var err error
	v := &jsonValue{}
	c := d.pop()
	switch c {
	case 'n':
		err = d.parseLiteral([]byte("null"), v, ValueNull)
	case 't':
		err = d.parseLiteral([]byte("true"), v, ValueTrue)
	case 'f':
		err = d.parseLiteral([]byte("false"), v, ValueFalse)
	case '"':
		err = d.parseString(v)
	case '[':
		err = d.parseArray(v)
	case '{':
		err = d.parseObject(v)
	default:
		err = d.parseNumber(v)
	}
	return v, err
}

// except 判断 byte 是否如期待的一样
func (d *jsonParse) except(a byte, b byte) error {
	if a != b {
		return d.error(a, fmt.Sprintf(" should be equal %s", string(b)))
	}
	return nil
}

func (d *jsonParse) equal(b byte) bool {
	if d.off > len(d.data)-1 {
		return false
	}
	return d.data[d.off] == b
}

func (d *jsonParse) next() byte {
	if d.off > len(d.data)-1 {
		return 0
	}
	d.off++
	return d.pop()
}

func (d *jsonParse) pop() byte {
	if d.off > len(d.data)-1 {
		return 0
	}
	return d.data[d.off]
}

func (d *jsonParse) skipWhiteSpace() {
	for ; d.equal(' ') || d.equal('\t') || d.equal('\n') || d.equal('\r'); d.off++ {
	}
}

func (d *jsonParse) parseLiteral(literal []byte, v *jsonValue, valueType ValueType) error {
	c := d.pop()
	if err := d.except(literal[0], c); err != nil {
		return err
	}
	c = d.next()
	for i := 1; i < len(literal); i++ {
		if c != literal[i] {
			return d.error(literal[i], fmt.Sprintf("parseJson type %d error", valueType))
		}
		c = d.next()
	}
	v.valueType = valueType
	return nil
}

func (d *jsonParse) parseNumber(v *jsonValue) error {
	start := d.off
	c := d.pop()
	// 判断负数
	if c == '-' {
		c = d.next()
	}
	// 判断整数部分
	if c == '0' {
		c = d.next()
	} else {
		if !isDigit1To9(c) {
			return d.error(c, "number syntax invalid")
		}
		c = d.next()
		for isDigit(c) {
			c = d.next()
		}
	}
	// 判断小数部分
	if c == '.' {
		c = d.next()
		if !isDigit(c) {
			return d.error(c, "number syntax invalid")
		}
		for isDigit(c) {
			c = d.next()
		}
	}
	// 判断科学计数部分
	if c == 'e' || c == 'E' {
		c = d.next()
		if c == '-' || c == '+' {
			c = d.next()
		}
		if !isDigit(c) {
			return d.error(c, "number syntax invalid")
		}
		for isDigit(c) {
			c = d.next()
		}
	}
	// 使用float64存储数字
	s := d.data[start:d.off]
	n, err := convertNumber(string(s))
	if err != nil {
		return err
	}
	v.n = n
	v.valueType = ValueNumber
	return nil
}

// 字符串的十六进制转化成十进制
func (d *jsonParse) parseHex4() (rune, error) {
	var r rune
	for i := 0; i < 4; i++ {
		c := d.next()
		r <<= 4
		if c >= '0' && c <= '9' {
			r |= rune(c - '0')
		} else if c >= 'A' && c <= 'F' {
			r |= rune(c - 'A' + 10)
		} else if c >= 'a' && c <= 'f' {
			r |= rune(c - 'a' + 10)
		} else {
			return 0, d.error(c, "invalid_string_unicode_hex")
		}
	}
	return r, nil
}

func (d *jsonParse) parseString(v *jsonValue) error {
	c := d.pop()
	if err := d.except(c, '"'); err != nil {
		return err
	}
	c = d.next()
	buf := bytes.NewBufferString("")
	for {
		switch c {
		case '"':
			v.s = buf.Bytes()
			c = d.next()
			v.valueType = ValueString
			return nil
		case '\\':
			c = d.next()
			switch c {
			case '"':
				buf.WriteByte('"')
			case '\\':
				buf.WriteByte('\\')
			case '/':
				buf.WriteByte('/')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'u':
				r, err := d.parseHex4()
				if err != nil {
					return err
				}
				if r >= 0xD800 && r <= 0xDFFF { // surrogate pair
					c = d.next()
					if c != '\\' {
						return d.error(c, "invalid_unicode_surrogate")
					}
					c = d.next()
					if c != 'u' {
						return d.error(c, "invalid_unicode_surrogate")
					}
					r2, err := d.parseHex4()
					if r2 < 0xDC00 || r2 > 0xDFFF {
						return d.error(c, "invalid_unicode_surrogate")
					}
					if err != nil {
						return err
					}
					r = (((r - 0xD800) << 10) | (r2 - 0xDC00)) + 0x10000
				}
				buf.WriteRune(r)
			default:
				return d.error(c, "invalid_string_escape")
			}
			c = d.next()
		case 0:
			return d.error(c, "miss quotation mark")
		default:
			if c < 0x20 {
				return d.error(c, "invalid string char")
			}
			buf.WriteByte(c)
			c = d.next()
		}
	}

}

func (d *jsonParse) parseArray(v *jsonValue) error {
	// 判断开头
	c := d.pop()
	if err := d.except(c, '['); err != nil {
		return err
	}
	d.next()
	d.skipWhiteSpace()
	if c = d.pop(); c == ']' {
		d.next()
		v.array.len = len(v.array.values)
		v.valueType = ValueArray
		return nil
	}
	for {
		c = d.pop()
		// 解析值
		d.skipWhiteSpace()
		v2, err := d.parserValue()
		if err != nil {
			return err
		}
		v.array.values = append(v.array.values, v2)
		d.skipWhiteSpace()
		c = d.pop()
		// 分析是否有分隔符
		if c == ',' {
			c = d.next()
		} else if c == ']' {
			d.next()
			v.array.len = len(v.array.values)
			v.valueType = ValueArray
			return nil
		} else {
			return d.error(c, "MISS_COMMA_OR_SQUARE_BRACKET")
		}
	}
}

func (d *jsonParse) parseObject(v *jsonValue) error {
	// 判断开头
	c := d.pop()
	if err := d.except(c, '{'); err != nil {
		return err
	}
	d.next()
	d.skipWhiteSpace()
	if c = d.pop(); c == '}' {
		d.next()
		v.object.size = len(v.object.values)
		v.valueType = ValueObject
		return nil
	}
	for {
		// 解析key
		d.skipWhiteSpace()
		key := &jsonValue{}
		if err := d.parseString(key); err != nil {
			return d.error(c, "miss key")
		}
		// 解析 ：字符
		d.skipWhiteSpace()
		c = d.pop()
		if c == ':' {
			c = d.next()
		} else {
			return d.error(c, "miss colon")
		}
		// 解析value
		d.skipWhiteSpace()
		v2, err := d.parserValue()
		if err != nil {
			return err
		}
		v.object.keys = append(v.object.keys, key)
		v.object.values = append(v.object.values, v2)

		// 解析分隔符、结束符
		d.skipWhiteSpace()
		c = d.pop()
		if c == ',' {
			c = d.next()
		} else if c == '}' {
			d.next()
			v.object.size = len(v.object.values)
			v.valueType = ValueObject
			return nil
		} else {
			return d.error(c, "miss comma or curly bracket")
		}
	}
}

func (d *jsonParse) error(c byte, context string) error {
	return &SyntaxError{msg: "invalid character " + string(c) + " " + context}
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isDigit1To9(c byte) bool {
	return c >= '1' && c <= '9'
}

func convertNumber(s string) (float64, error) {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, &UnmarshalTypeError{Value: "number out of range " + s, Type: reflect.TypeOf(0.0)}
	}
	return n, err
}
