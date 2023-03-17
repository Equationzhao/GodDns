/*
 *     @Copyright
 *     @file: Util.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/18 上午12:59
 *     @last modified: 2023/3/17 下午11:42
 *
 *
 *
 */

package Util

import (
	"fmt"
	"reflect"
	"runtime"
)

// OSDetect OS detection
func OSDetect() string {
	return runtime.GOOS
}

// ConvertableKeyValue  Convert any type to key-value according to the format
type ConvertableKeyValue interface {
	Convert2KeyValue(format string) string
}

// Convert2KeyValue Convert any type to key-value according to the format
// if the type implements ConvertableKeyValue, use its method
// else use reflection
// if the type has a field named "KeyValue", use it as key
// else if the type has a field named "json", use it as key
// else use the field name as key
// example:
//
//		type A struct {
//			Device string `KeyValue:"device" json:"device"`
//			IP     string `json:"ip"`
//		 	Type   string
//	     	unexported string
//		}
//
// a := A{Device: "device", IP: "ip", Type: "type"}
//
// fmt.Println(Convert2KeyValue("%s: %s", a))
//
// output:
//
// device: device
//
// ip: ip
//
// Type: type
func Convert2KeyValue(format string, i any) string {

	if _, ok := i.(ConvertableKeyValue); ok {
		return i.(ConvertableKeyValue).Convert2KeyValue(format)
	}

	var content string
	v := reflect.ValueOf(i)
	t := reflect.TypeOf(i)
	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			continue
		}
		name := t.Field(i).Tag.Get("KeyValue")
		if name == "" {
			name = t.Field(i).Tag.Get("json")
			if name == "" {
				name = t.Field(i).Name
			}
		}
		content += fmt.Sprintf(format, name, v.Field(i).Interface()) + "\n"
	}

	return content
}

type ConvertableXWWWFormUrlencoded interface {
	Convert2XWWWFormUrlencoded() string
}

// Convert2XWWWFormUrlencoded Convert to x-www-form-urlencoded format
// if the type implements ConvertableXWWWFormUrlencoded, use its method
// else use reflection
// if the type has a field named "xwwwformurlencoded", use it as key
// else if the type has a field named "json", use it as key
// else use the field name as key
// example:
//
//		type A struct {
//			Device string `xwwwformurlencoded:"device" json:"device"`
//			IP     string `json:"ip"`
//		 	Type   string
//	     	unexported string
//		}
//
// a := A{Device: "device", IP: "ip", Type: "type"}
//
// fmt.Println(Convert2XWWWFormUrlencoded(a))
//
// output:
//
// device=device&ip=ip&Type=type
func Convert2XWWWFormUrlencoded(i any) string {

	if _, ok := i.(ConvertableXWWWFormUrlencoded); ok {
		return i.(ConvertableXWWWFormUrlencoded).Convert2XWWWFormUrlencoded()
	}

	v := reflect.ValueOf(i)

	var content string
	t := reflect.TypeOf(i)
	n := t.NumField()
	for i := 0; i < n; i++ {
		if !t.Field(i).IsExported() {
			if i == n-1 && content != "" {
				content = content[:len(content)-1]
			}

			continue
		}
		name := t.Field(i).Tag.Get("xwwwformurlencoded")
		if name == "" {
			// if there's no "xwwwformurlencoded" Tag, use "json" instead
			name = t.Field(i).Tag.Get("json")
			// if there's no "json" Tag, use field name instead
			// ? how to deal with json include string "omitempty"
			if name == "" {
				name = t.Field(i).Name
			}
		}
		content += fmt.Sprintf("%+v", name) + "=" + fmt.Sprintf("%v", v.Field(i).Interface())
		if i != n-1 {
			content += "&"
		}

	}

	return content
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
//			Msg:     "Hello",
//			Success: DDNS.Success,
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

type Pair[T, U any] struct {
	First  T
	Second U
}

func (receiver *Pair[T, U]) Set(first T, second U) {
	receiver.First = first
	receiver.Second = second
}
