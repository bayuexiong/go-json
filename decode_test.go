package json

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

/**
 * 1. 判断类型
 * 2. 判断值
 * 3. 判断错误
 */
func parseJson(t *testing.T, data []byte) (*jsonValue, error) {
	t.Helper()
	var decode *jsonParse
	decode = new(jsonParse)
	decode.init(data)
	value, err := decode.parser()
	if err != nil {
		t.Errorf("parseJson %s error %s", string(data), err.Error())
		return nil, err
	}
	return value, nil
}

func assertTrue(t *testing.T, value bool) {
	t.Helper()
	if !value {
		t.Errorf("Should be True, But %v", value)
	}
}

func assertFalse(t *testing.T, value bool) {
	t.Helper()
	if value {
		t.Errorf("Should be True, But %v", value)
	}
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Should be equal, But expected is %s, actual is %s", expected, actual)
	}
}

func equalError(a, b error) bool {
	if a == nil {
		return b == nil
	}
	if b == nil {
		return a == nil
	}
	return a.Error() == b.Error()
}

func testError(t *testing.T, data []byte, msg string) {
	t.Helper()
	var decode *jsonParse
	decode = new(jsonParse)
	decode.init(data)
	_, err := decode.parser()
	if err == nil {
		t.Errorf("data %s should be error, but pass", data)
		return
	}

	if !strings.Contains(err.Error(), msg) {
		t.Errorf("Data %s Should be error is [%s], but error is [%s]", data, msg, err.Error())
	}

}

func TestParseNull(t *testing.T) {
	value, err := parseJson(t, []byte("null"))
	if err != nil {
		return
	}
	assertTrue(t, value.getValueType() == ValueNull)
}

func TestParseTrue(t *testing.T) {
	value, err := parseJson(t, []byte("true"))
	if err != nil {
		return
	}
	assertTrue(t, value.getValueType() == ValueTrue)
}

func TestParseFalse(t *testing.T) {
	value, err := parseJson(t, []byte("false"))
	if err != nil {
		return
	}
	assertTrue(t, value.getValueType() == ValueFalse)
}

func TestParseInvalidValue(t *testing.T) {
	testError(t, []byte("nul"), fmt.Sprintf("invalid character l parseJson type %d error", ValueNull))
	testError(t, []byte("?"), "invalid character ? number syntax invalid")
	testError(t, []byte("null x"), "unexpected end of JSON input")
	/* invalid number */
	testError(t, []byte("+0"), "number syntax invalid")
	testError(t, []byte("+1"), "number syntax invalid")
	testError(t, []byte(".123"), "number syntax invalid") /* at least one digit before '.' */
	testError(t, []byte("1."), "number syntax invalid")   /* at least one digit after '.' */
	testError(t, []byte("INF"), "number syntax invalid")
	testError(t, []byte("inf"), "number syntax invalid")
	testError(t, []byte("NAN"), "number syntax invalid")
	testError(t, []byte("nan"), "invalid character u parseJson type 0 error")
	testError(t, []byte("0123"), "unexpected end of JSON input")
	testError(t, []byte("0x0"), "unexpected end of JSON input")
	testError(t, []byte("0x123"), "unexpected end of JSON input")
	// invalid string
	testError(t, []byte("\""), "miss quotation mark")
	testError(t, []byte("\"abc"), "miss quotation mark")
	testError(t, []byte("\"\x01\""), "invalid string char")
	testError(t, []byte("\"\x1f\""), "invalid string char")
	// invalid array
	testError(t, []byte("[1,]"), "invalid character")
	testError(t, []byte("[\"a\",nul]"), "invalid character")

}

func testString(t *testing.T, expect, source string) {
	t.Helper()
	value, err := parseJson(t, []byte(source))
	if err != nil {
		return
	}
	assertTrue(t, value.getValueType() == ValueString)
	v, err := value.getString()
	if err != nil {
		t.Errorf("get jsonValue error %source", err.Error())
	}
	assertEqual(t, v, expect)

}

func TestParseString(t *testing.T) {
	testString(t, "", "\"\"")
	testString(t, "Hello", "\"Hello\"")
	testString(t, "Hello\nWorld", "\"Hello\\nWorld\"")
	testString(t, "\" \\ / \b \f \n \r \t", "\"\\\" \\\\ \\/ \\b \\f \\n \\r \\t\"")
	testString(t, "$", "\"\\u0024\"")        /* Dollar sign U+0024 */
	testString(t, "¢", "\"\\u00A2\"")        /* Cents sign U+00A2 */
	testString(t, "€", "\"\\u20AC\"")        /* Euro sign U+20AC */
	testString(t, "𝄞", "\"\\uD834\\uDD1E\"") /* G clef sign U+1D11E */
	testString(t, "𝄞", "\"\\ud834\\udd1e\"") /* G clef sign U+1D11E */
}

func TestParseInvalidUnicodeHex(t *testing.T) {
	testError(t, []byte("\"\\u\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u0\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u01\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u012\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u/000\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\uG000\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u0/00\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u0G00\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u0/00\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u000/\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u000G\""), "invalid_string_unicode_hex")
	testError(t, []byte("\"\\u 123\""), "invalid_string_unicode_hex")
}

func TestParseInvalidUnicodeSurrogate(t *testing.T) {
	testError(t, []byte("\"\\uD800\""), "invalid_unicode_surrogate")
	testError(t, []byte("\"\\uDBFF\""), "invalid_unicode_surrogate")
	testError(t, []byte("\"\\uD800\\\\\""), "invalid_unicode_surrogate")
	testError(t, []byte("\"\\uD800\\uDBFF\""), "invalid_unicode_surrogate")
	testError(t, []byte("\"\\uD800\\uE000\""), "invalid_unicode_surrogate")

}

func testNumber(t *testing.T, expect float64, source string) {
	t.Helper()
	v, err := parseJson(t, []byte(source))
	if err != nil {
		return
	}
	assertTrue(t, v.getValueType() == ValueNumber)
	n, _ := v.getNumber()
	assertEqual(t, expect, n)
}

func TestParseNumber(t *testing.T) {
	testNumber(t, 0.0, "0")
	testNumber(t, 0.0, "-0")
	testNumber(t, 0.0, "-0.0")
	testNumber(t, 1.0, "1")
	testNumber(t, -1.0, "-1")
	testNumber(t, 1.5, "1.5")
	testNumber(t, -1.5, "-1.5")
	testNumber(t, 3.1416, "3.1416")
	testNumber(t, 1e10, "1E10")
	testNumber(t, 1e10, "1e10")
	testNumber(t, 1e+10, "1E+10")
	testNumber(t, 1e-10, "1E-10")
	testNumber(t, -1e10, "-1E10")
	testNumber(t, -1e10, "-1e10")
	testNumber(t, -1e+10, "-1E+10")
	testNumber(t, -1e-10, "-1E-10")
	testNumber(t, 1.234e+10, "1.234E+10")
	testNumber(t, 1.234e-10, "1.234E-10")
	testNumber(t, 0.0, "1e-10000") /* must underflow */

	testNumber(t, 1.0000000000000002, "1.0000000000000002")           /* the smallest number > 1 */
	testNumber(t, 4.9406564584124654e-324, "4.9406564584124654e-324") /* minimum denormal */
	testNumber(t, -4.9406564584124654e-324, "-4.9406564584124654e-324")
	testNumber(t, 2.2250738585072009e-308, "2.2250738585072009e-308") /* Max subnormal double */
	testNumber(t, -2.2250738585072009e-308, "-2.2250738585072009e-308")
	testNumber(t, 2.2250738585072014e-308, "2.2250738585072014e-308") /* Min normal positive double */
	testNumber(t, -2.2250738585072014e-308, "-2.2250738585072014e-308")
	testNumber(t, 1.7976931348623157e+308, "1.7976931348623157e+308") /* Max double */
	testNumber(t, -1.7976931348623157e+308, "-1.7976931348623157e+308")
}

func TestOverflowNumber(t *testing.T) {
	testError(t, []byte("1e309"), "number out of range")
	testError(t, []byte("-1e309"), "number out of range")
}

func TestParseArray(t *testing.T) {
	var (
		v   *jsonValue
		err error
	)
	v, err = parseJson(t, []byte("[ ]"))
	if err != nil {
		return
	}
	assertTrue(t, v.getValueType() == ValueArray)
	assertEqual(t, v.getArrayLen(), 0)

	v, err = parseJson(t, []byte("[1,2,3]"))
	if err != nil {
		return
	}
	assertTrue(t, v.getValueType() == ValueArray)
	assertEqual(t, v.getArrayLen(), 3)

	func() {
		v1, _ := v.getArrayElem(0)
		assertTrue(t, v1.valueType == ValueNumber)
		n1, _ := v1.getNumber()
		assertEqual(t, n1, 1.0)

		v2, _ := v.getArrayElem(1)
		assertTrue(t, v2.valueType == ValueNumber)
		n2, _ := v2.getNumber()
		assertEqual(t, n2, 2.0)

		v3, _ := v.getArrayElem(2)
		assertTrue(t, v3.valueType == ValueNumber)
		n3, _ := v3.getNumber()
		assertEqual(t, n3, 3.0)
	}()

	v, err = parseJson(t, []byte("[\"a\", \"bb\", \"ccc\"]"))
	if err != nil {
		return
	}
	assertTrue(t, v.getValueType() == ValueArray)
	assertEqual(t, v.getArrayLen(), 3)
	func() {
		v1, _ := v.getArrayElem(0)
		assertTrue(t, v1.valueType == ValueString)
		s1, _ := v1.getString()
		assertEqual(t, s1, "a")

		v2, _ := v.getArrayElem(1)
		assertTrue(t, v2.valueType == ValueString)
		s2, _ := v2.getString()
		assertEqual(t, s2, "bb")

		v3, _ := v.getArrayElem(2)
		assertTrue(t, v3.valueType == ValueString)
		s3, _ := v3.getString()
		assertEqual(t, s3, "ccc")
	}()

	v, err = parseJson(t, []byte("[ [ ] , [ 0 ] , [ 0 , 1 ] , [ 0 , 1 , 2 ] ]"))
	if err != nil {
		return
	}
	assertTrue(t, v.getValueType() == ValueArray)
	assertEqual(t, v.getArrayLen(), 4)

	func() {
		v1, _ := v.getArrayElem(0)
		assertTrue(t, v1.valueType == ValueArray)

		// [0]
		v2, _ := v.getArrayElem(1)
		assertTrue(t, v2.valueType == ValueArray)
		a1, _ := v2.getArrayElem(0)
		n1, _ := a1.getNumber()
		assertEqual(t, n1, 0.0)

		// [0,1]
		v3, _ := v.getArrayElem(2)
		assertTrue(t, v3.valueType == ValueArray)
		a2, _ := v3.getArrayElem(1)
		n21, _ := a2.getNumber()
		assertEqual(t, n21, 1.0)
	}()
}

func TestMissComaOrSquareBracket(t *testing.T) {
	testError(t, []byte("[1"), "MISS_COMMA_OR_SQUARE_BRACKET")
	testError(t, []byte("[1}"), "MISS_COMMA_OR_SQUARE_BRACKET")
	testError(t, []byte("[1 2"), "MISS_COMMA_OR_SQUARE_BRACKET")
	testError(t, []byte("[[]"), "MISS_COMMA_OR_SQUARE_BRACKET")
}

func TestParseObject(t *testing.T) {
	var (
		v   *jsonValue
		err error
	)
	v, err = parseJson(t, []byte("{ }"))
	if err != nil {
		return
	}
	assertTrue(t, v.getValueType() == ValueObject)
	assertEqual(t, v.getObjectSize(), 0)

	v, err = parseJson(t, []byte(` { 
	"n" : null , 
	"f" : false , 
	"t" : true , 
	"i" : 123 , 
	"s" : "abc", 
	"a" : [ 1, 2, 3 ],
	"o" : { "1" : 1, "2" : 2, "3" : 3 }
	 } `))
	if err != nil {
		return
	}
	assertTrue(t, v.getValueType() == ValueObject)
	assertEqual(t, v.getObjectSize(), 7)
	func() {
		k1, _ := v.getObjectKey(0)
		assertTrue(t, k1.getValueType() == ValueString)
		kv1, _ := k1.getString()
		assertEqual(t, kv1, "n")
		v1, _ := v.getObjectValue(0)
		assertTrue(t, v1.getValueType() == ValueNull)

		v4, _ := v.getObjectValue(3)
		assertTrue(t, v4.getValueType() == ValueNumber)
		vv4, _ := v4.getNumber()
		assertEqual(t, vv4, 123.0)

		v5, _ := v.getObjectValue(4)
		assertTrue(t, v5.getValueType() == ValueString)
		vv5, _ := v5.getString()
		assertEqual(t, vv5, "abc")

		arr, _ := v.getObjectValue(5)
		assertTrue(t, arr.getValueType() == ValueArray)
		assertEqual(t, arr.getArrayLen(), 3)
		a1, _ := arr.getArrayElem(0)
		av1, _ := a1.getNumber()
		assertEqual(t, av1, 1.0)
		a2, _ := arr.getArrayElem(1)
		av2, _ := a2.getNumber()
		assertEqual(t, av2, 2.0)

		obj, _ := v.getObjectValue(6)
		assertTrue(t, obj.getValueType() == ValueObject)
		assertEqual(t, obj.getObjectSize(), 3)

		o1, _ := obj.getObjectValue(0)
		assertTrue(t, o1.getValueType() == ValueNumber)
		ov1, _ := o1.getNumber()
		assertEqual(t, ov1, 1.0)

		o2, _ := obj.getObjectValue(2)
		assertTrue(t, o2.getValueType() == ValueNumber)
		ov2, _ := o2.getNumber()
		assertEqual(t, ov2, 3.0)

	}()
}

func TestParseMissKey(t *testing.T) {
	testError(t, []byte("{:1,"), "miss key")
	testError(t, []byte("{1:1,"), "miss key")
	testError(t, []byte("{true:1"), "miss key")
	testError(t, []byte("{false:1"), "miss key")
	testError(t, []byte("{null:1"), "miss key")
	testError(t, []byte("{[]:1"), "miss key")
	testError(t, []byte("{{}:1"), "miss key")
	testError(t, []byte("{\"a\":1,"), "miss key")
}

func TestParseMissColon(t *testing.T) {
	testError(t, []byte("{\"a\"}"), "miss colon")
	testError(t, []byte("{\"a\", \"b\"}"), "miss colon")
}

func TestMissComaOrCurlyBracket(t *testing.T) {
	testError(t, []byte("{\"a\": 1"), "miss comma or curly bracket")
	testError(t, []byte("{\"a\": 1]"), "miss comma or curly bracket")
	testError(t, []byte("{\"a\": 1 \"b\""), "miss comma or curly bracket")
	testError(t, []byte("{\"a\": {}"), "miss comma or curly bracket")
}
