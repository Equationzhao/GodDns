package Util

import (
	"fmt"
	"net/url"
	"reflect"
	"runtime"
	"strings"
)

// OSDetect OS detection
// return os and arch
// if it's 386, replace it with x86
func OSDetect() (os string, arch string) {
	os = runtime.GOOS
	arch = runtime.GOARCH
	switch arch {
	case "amd64":
	case "386":
		arch = "x86"
	default:

	}

	return
}

// ConvertableKeyValue  Convert any type to key-value according to the format
type ConvertableKeyValue interface {
	Convert2KeyValue(format string) string
}

// Convert2KeyValue Convert any type to key-value according to the format
// if the type implements ConvertableKeyValue, use its method
// else use reflection
// if the type has a tag named "KeyValue", use it as key (the first word is key, the rest content after the first ',' is comments)
// else if the type has a tag named "json", use it as key (use the first word before the first ',' as key)
// else use the field name as key
// example:
//
//	type B struct {
//		X string
//		x string
//	}
//
//	type A struct {
//		Device     string `KeyValue:"device,device name" json:"device"`
//		IP         string `json:"ip,omitempty,string"`
//		Type       string
//		unexported string
//		B          B
//	}
//		a := A{Device: "device", IP: "ip", Type: "type", B: B{X: "123", x: "321"}}
//		fmt.Println(Convert2KeyValue("%s: %s", a))
//		output:
//	 # device name
//		device: device
//		ip: ip
//		Type: type
//		B: {123 321}
func Convert2KeyValue(format string, i any) string {

	if _, ok := i.(ConvertableKeyValue); ok {
		return i.(ConvertableKeyValue).Convert2KeyValue(format)
	}

	var content strings.Builder
	v := reflect.ValueOf(i)
	t := reflect.TypeOf(i)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		tfieldi := t.Field(i)
		if !tfieldi.IsExported() {
			continue
		}

		vfiledi := v.Field(i).Interface()
		if reflect.DeepEqual(vfiledi, reflect.Zero(reflect.TypeOf(vfiledi)).Interface()) {
			continue
		}

		name := tfieldi.Tag.Get("KeyValue") // `name,comments`
		if name == "-" {
			continue
		}

		comments := ""

		if strings.Contains(name, ",") {
			comments = strings.SplitN(name, ",", 2)[1] // get comments
			name = strings.Split(name, ",")[0]         // get name
		}

		if name == "" {
			name = tfieldi.Tag.Get("json")
			if strings.Contains(name, ",") {
				name = strings.SplitN(name, ",", 2)[0] // get name, remove ",omitempty"
			}
		}

		if name == "-" {
			continue
		}

		if name == "" {
			name = t.Field(i).Name
		}

		if comments != "" {
			content.WriteString(fmt.Sprintf(" # %s", comments))
			content.WriteByte('\n')
			content.WriteString(fmt.Sprintf(format, name, vfiledi))
			content.WriteByte('\n')
		} else {
			content.WriteString(fmt.Sprintf(format, name, vfiledi))
			content.WriteByte('\n')
		}

	}
	return content.String()
}

// ConvertableXWWWFormUrlencoded Convert any type to x-www-form-urlencoded format
type ConvertableXWWWFormUrlencoded interface {
	Convert2XWWWFormUrlencoded() string
}

// Convert2XWWWFormUrlencoded Convert to x-www-form-urlencoded format
// if the type implements ConvertableXWWWFormUrlencoded, use its method
// else use reflection
// if the type has a field named "xwwwformurlencoded", use it as key
// else if the type has a field named "json", use it as key
// else use the field name as key
//
// // example:
//
//		type A struct {
//			Device string `xwwwformurlencoded:"device" json:"device"`
//			IP     string `json:"ip"`
//		 	Type   string
//		    unexported string
//		}
//		a := A{Device: "device", IP: "ip", Type: "type"}
//		fmt.Println(Convert2XWWWFormUrlencoded(a))
//
//			output:device=device&ip=ip&Type=typ
//
//		type B struct {
//		        X string
//		        x string
//		}
//
//		ab:=struct {
//			A
//			B
//		}{A: a, B: B{X: "123", x: "321"}}
//
//		fmt.Println(Convert2XWWWFormUrlencoded(ab))
//
//
//	 	output:device=device&ip=ip&Type=type&X=123
//
//		m := map[string]string{"device": "device", "ip": "ip", "Type": "type"}
//		fmt.Println(Convert2XWWWFormUrlencoded(m))
//		output:
//		device=device&ip=ip&Type=type
//
//		s := []string{"device", "ip", "type"}
//		fmt.Println(Convert2XWWWFormUrlencoded(s))
//
//	   	output:=device&=ip&=type
//
// ! [need test]
func Convert2XWWWFormUrlencoded(i any) string {
	return convert2xwwwformurlencoded(i, true)
}

func convert2xwwwformurlencoded(i any, isTheLast bool) string {
	if i == nil {
		return ""
	}

	if c, ok := i.(ConvertableXWWWFormUrlencoded); ok {
		return c.Convert2XWWWFormUrlencoded()
	}

	v := reflect.ValueOf(i)
	t := reflect.TypeOf(i)

	var content strings.Builder

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		if isTheLast {
			// the last element
			content.WriteByte('=')
			content.WriteString(url.QueryEscape(v.String()))
			return content.String()
		} else {
			// not the last element
			content.WriteByte('=')
			content.WriteString(url.QueryEscape(v.String()))
			content.WriteByte('&')
			return content.String()
		}
	case reflect.Map:
		iter := v.MapRange()
		l := v.Len()
		for iter.Next() {
			l--
			v := iter.Value()
			t := v.Kind()
			if t == reflect.Pointer {
				v = v.Elem()
				t = iter.Value().Elem().Kind()
			}
			// if v is not struct/map/pointer
			switch t {
			case reflect.Struct:
				fallthrough
			case reflect.Map:
				if l > 0 {
					// not the last element
					content.WriteString(convert2xwwwformurlencoded(iter.Value(), false))
				} else {
					content.WriteString(convert2xwwwformurlencoded(iter.Value(), true))
				}
			case reflect.Interface:
				k := iter.Key()
				switch v.Interface().(type) {
				// basic type
				case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128, bool, []byte, []rune, uintptr, nil:
					if l > 0 {
						// not the last element
						// content.WriteString(fmt.Sprintf("%s=%s&", k, url.QueryEscape(fmt.Sprint(v))))
						content.WriteString(k.String())
						content.WriteByte('=')
						content.WriteString(url.QueryEscape(fmt.Sprint(v)))
						content.WriteByte('&')
					} else {
						// content.WriteString(fmt.Sprintf("%s=%s", k, url.QueryEscape(fmt.Sprint(v))))
						content.WriteString(k.String())
						content.WriteByte('=')
						content.WriteString(url.QueryEscape(fmt.Sprint(v)))
					}
				default:
					if l > 0 {
						// not the last element
						content.WriteString(convert2xwwwformurlencoded(v.Interface(), false))
					} else {
						content.WriteString(convert2xwwwformurlencoded(v.Interface(), true))
					}
				}
			default:
				k := iter.Key()
				if l > 0 {
					// not the last element
					content.WriteString(fmt.Sprintf("%s=%s&", k, url.QueryEscape(fmt.Sprint(v))))
				} else {
					content.WriteString(fmt.Sprintf("%s=%s", k, url.QueryEscape(fmt.Sprint(v))))
				}
			}
		}
	case reflect.Struct:
		n := t.NumField()
		pieces := make([]string, 0, n)
		for i := 0; i < n; i++ {
			fieldi := t.Field(i)
			if !fieldi.IsExported() {
				continue
			}
			name := fieldi.Tag.Get("xwwwformurlencoded")
			if name == "-" {
				continue
			}
			if name == "" {
				// if there's no "xwwwformurlencoded" Tag, use "json" instead
				name = fieldi.Tag.Get("json")

				if name == "-" {
					continue
				}

				if strings.Contains(name, ",") {
					name = strings.SplitN(name, ",", 2)[0] // remove ",omitempty"
				}
				// if there's no "json" Tag, use field name instead
				if name == "" {
					name = fieldi.Name
				}
			}

			fieldType := fieldi.Type
			if fieldType.Kind() == reflect.Struct {
				pieces = append(pieces, convert2xwwwformurlencoded(v.Field(i).Interface(), true))
				continue
			}

			pieces = append(pieces, fmt.Sprintf("%s=%s", url.QueryEscape(name), url.QueryEscape(fmt.Sprintf("%v", v.Field(i).Interface()))))
		}
		content.WriteString(strings.Join(pieces, "&"))

		if !isTheLast && content.Len() != 0 {
			content.WriteByte('&')
		}

	case reflect.Slice, reflect.Array:
		// if the slice is empty, return ""
		l := v.Len()
		if l == 0 {
			return ""
		}

		for i := 0; i < l-1; i++ {
			content.WriteString(convert2xwwwformurlencoded(v.Index(i).Interface(), false))
		}
		content.WriteString(convert2xwwwformurlencoded(v.Index(l-1).Interface(), isTheLast))
	}

	return content.String()
}

// HasVariable Check if struct has variable by name
// can access variable both exported and unexported
// if `i` is a pointer, it will be dereferenced
func HasVariable(i any, name string) bool {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		return v.FieldByName(name).IsValid()
	} else {
		// not a struct
		return false
	}
}

// GetVariable
// Get variable from struct by name
// can only get exported variable
// if `i` is a pointer, it will be dereferenced
func GetVariable(i any, name string) (any, error) {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {

		// check whether the struct has the variable && the variable is exported
		if v.FieldByName(name).IsValid() {

			// check whether the variable is exported
			if v.FieldByName(name).CanInterface() {
				return v.FieldByName(name).Interface(), nil
			} else {
				return nil, fmt.Errorf("field `%s` is unexported", name)
			}

		} else {
			return nil, fmt.Errorf("no such field `%s`", name)
		}

	} else {
		// not a struct
		return nil, fmt.Errorf("not a struct")
	}
}

// SetVariable Set Field from struct by name
// can only set exported variable
// ptr2i must be a pointer to struct otherwise it will return an error because the field can't be set
func SetVariable(ptr2i any, name string, value any) error {

	v := reflect.ValueOf(ptr2i)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	} else {
		return fmt.Errorf("not a pointer")
	}

	if v.Kind() == reflect.Struct {
		// check whether the struct has the variable && the variable is exported
		if v.FieldByName(name).IsValid() {
			// check whether the variable is settable
			if v.FieldByName(name).CanSet() {
				// check whether the variable type is the same as the value type
				if v.FieldByName(name).Type() == reflect.TypeOf(value) {
					v.FieldByName(name).Set(reflect.ValueOf(value))
					return nil

				} else {
					return fmt.Errorf("type of value to set is not the same as the type of field `%s`", name)
				}
			} else {
				return fmt.Errorf("field `%s` is unexported", name)
			}

		} else {
			return fmt.Errorf("no such field `%s`", name)
		}
	} else {
		// not a struct
		return fmt.Errorf("not a struct")
	}

}

// GetTypeName Get type name of variable
// example:
//
//	  	s := DDNS.Status{
//			Name:    "Test",
//			MG:     "Hello",
//			Status: DDNS.Status,
//	 	}
//
//		fmt.(Util.GetTypeName(s))
//		fmt.(Util.GetTypeName(&s))
//
//		b := make(map[string]int, 10)
//		c := make([]string, 10)
//
//		fmt.(Util.GetTypeName(b))
//		fmt.(Util.GetTypeName(c))
//
// output:
//
//	DDNS.Status
//	*DDNS.Status
//	map[string]int
//	[]string
func GetTypeName(variable any) string {
	return reflect.TypeOf(variable).String()
}
