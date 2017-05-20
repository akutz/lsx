package lsx

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	configScopeKey  = `@scope@`
	configParentKey = `@parent@`
)

// Config is the type used to pass around configuration information.
//
// Please use a Config object's Len function to get an accurate number
// of keys. Using len(c) will reflect a technically correct number, but
// will also include some metadata-specific keys used to track the
// Config object's scope information.
type Config map[string]interface{}

// Len returns the number of keys in the Config map less the
// meatadata-specific keys.
func (c Config) Len() int {
	l := len(c)
	if _, ok := c[configScopeKey]; ok {
		l--
	}
	if _, ok := c[configParentKey]; ok {
		l--
	}
	return l
}

// Parent returns the parent of this Config instance. If the Config
// is the ancestor instance, it will have no parent and nil will be
// returned.
func (c Config) Parent(ctx context.Context) Config {
	pmap, ok := c[configParentKey]
	if !ok {
		return nil
	}
	switch tpmap := pmap.(type) {
	case Config:
		return tpmap
	case map[string]interface{}:
		return Config(tpmap)
	}
	return nil
}

// Scope returns a scoped version of the config instance.
//
// The scope parameter adheres to a JSON path, dot-style notation.
// A scope's path tokens can also refer to array elements if those
// elements are themselves JSON objects. Instead of referring to
// these elements by index, the path token is used to to match the
// "name" field of an array element when that element is a JSON object.
func (c Config) Scope(ctx context.Context, scope string) Config {
	v := c.get(ctx, scope, false)
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	config := Config(m)
	config[configParentKey] = c
	config[configScopeKey] = scope
	return config
}

// Get returns a value from the config map.
//
// The path parameter adheres to a JSON path, dot-style notation.
// A path's tokens can also refer to array elements if those
// elements are themselves JSON objects. Instead of referring to
// these elements by index, the path token is used to to match the
// "name" field of an array element when that element is a JSON object.
func (c Config) Get(ctx context.Context, path string) interface{} {
	return c.get(ctx, path, true)
}
func (c Config) get(
	ctx context.Context,
	path string,
	askParent bool) interface{} {

	// if there is an environment variable set that matches the property
	// path, return the environment variable's value (if it's not empty)
	if v := os.Getenv(getEnvVarName(path)); v != "" {
		return v
	}

	var (
		// cur is the cursor pointing to the object that matches
		// the path token in the for loop below
		cur    interface{} = c
		tokens             = strings.Split(path, ".")
	)

	// iterate through the path tokens
	for tokIdx, tok := range tokens {

		var (
			// next points to the object that the cursor should point
			// at the end of each loop iteration
			next interface{}

			// curVal is the reflected value of the cursor
			curVal reflect.Value

			// isFinalToken is true when the loop is in its final iteration
			_ = tokIdx == len(tokens)-1
		)

		// get the cursor's reflected value
		curVal = reflect.ValueOf(cur)

		// dereference the cursor's reflected value
		curVal = derefValue(curVal)

		// isCurValNillable is a flag that indicates whether the current
		// value's type is nillable and isCurValNil is a flag that
		// indicates whether the current value is nil
		_, _ = isNillable(curVal)

		logCurVal(path, tokIdx, tok, curVal)

		switch curVal.Kind() {

		// if the map has a key that matches the path token then
		// assign the value for the key to next
		case reflect.Map:
			for _, mapKey := range curVal.MapKeys() {

				// get a string representation of the dereferenced map key
				szMapKey := toString(derefValue(mapKey).Interface())
				//fmt.Fprintf(os.Stderr, "szMapKey=%s\n", szMapKey)

				// if the map key's string representation matches the
				// path token then assign the value for the key to
				// next and break out of the immediate loop
				if strings.EqualFold(tok, szMapKey) {
					next = curVal.MapIndex(mapKey).Interface()
					break
				}
			}

		// iterate the array looking for maps and structs:
		//
		//        map      if the map has a key called "name" with a value
		//                 that matches the path token, assign the map
		//                 to next.
		//
		//        struct   if the struct has a field with a name that
		//                 matches the path token, assign the field's
		//                 value to next
		case reflect.Array, reflect.Slice:
			for x := 0; x < curVal.Len(); x++ {
				curValEl := curVal.Index(x)
				curValEl = derefValue(curValEl)

				logCurVal(path, tokIdx, tok, curValEl)

				switch curValEl.Kind() {
				// if the map has a key called "name" with a value
				// that matches the path token, assign the map
				// to next.
				case reflect.Map:
					for _, mapKey := range curValEl.MapKeys() {

						// get a string representation of the dereferenced
						// map key
						szMapKey := toString(derefValue(mapKey).Interface())
						//fmt.Fprintf(os.Stderr, "szMapKey=%s\n", szMapKey)

						// if the map key's string representation does not
						// match "name" then break ouf of the immediate loop
						if !strings.EqualFold("name", szMapKey) {
							continue
						}

						// the map key's string representation matches name,
						// so see if the value for the key matches the
						// path token
						mapVal := derefValue(curValEl.MapIndex(mapKey))
						szMapVal := toStringWithOpts(mapVal.Interface(), false)
						if strings.EqualFold(tok, szMapVal) {
							next = curValEl.Interface()
							break
						}
					}
				// if the struct has a field with a name that
				// matches the path token, assign the field's
				// value to next
				case reflect.Struct:
					curType := curValEl.Type()
					for fldIdx := 0; fldIdx < curType.NumField(); fldIdx++ {
						fldType := curType.Field(fldIdx)
						fldName := fldType.Name
						if strings.EqualFold("name", fldName) {
							fldVal := curValEl.Field(fldIdx)
							if fldVal.Kind() == reflect.String {
								szFldVal := fldVal.String()
								if strings.EqualFold(tok, szFldVal) {
									next = curValEl.Interface()
									break
								}
							}
						} else if strings.EqualFold(tok, fldName) {
							fldVal := curValEl.Field(fldIdx)
							next = fldVal.Interface()
							break
						}
					}
				}
			}

		// if the struct has a field with a name that matches
		// the path token assign the field's value to next
		case reflect.Struct:
			curType := curVal.Type()
			for fldIdx := 0; fldIdx < curType.NumField(); fldIdx++ {
				fldType := curType.Field(fldIdx)
				fldName := fldType.Name
				/*if strings.EqualFold("name", fldName) {
					fldVal := curVal.Field(fldIdx)
					if fldVal.Kind() == reflect.String {
						szFldVal := fldVal.String()
						if strings.EqualFold(tok, szFldVal) {
							next = curVal.Interface()
							break
						}
					}
				} else */if strings.EqualFold(tok, fldName) {
					fldVal := curVal.Field(fldIdx)
					next = fldVal.Interface()
					break
				}
			}
		}

		// update the cursor, and if the cursor is nil then go ahead
		// and break out of the immediate loop
		if cur = next; cur == nil {
			break
		}
	}

	// if the cursor is not nil then return its value
	if cur != nil {
		return cur
	}

	// since the cursor is known to be nil at this point, the decision is
	// now whether or not to query the Config instance's parent for the
	// same path

	// a false askParent flag explicitly determines to *not* query the parent
	if !askParent {
		return nil
	}

	// check for the Config instance's parent, and if it is not nil query it
	// for the propert path
	if parent := c.Parent(ctx); parent != nil {
		return parent.get(ctx, path, askParent)
	}

	// the property path was not found in this Config instance or in any
	// ancestoral Config instances, thus return nil
	return nil
}

var debug, _ = strconv.ParseBool(os.Getenv("LSX_DEBUG"))

func logCurVal(path string, tokIdx int, tok string, curVal reflect.Value) {
	if !debug {
		return
	}
	buf, _ := json.MarshalIndent(curVal.Interface(), "\t\t\t", "  ")
	fmt.Fprintf(os.Stderr, `
curVal:
	path		%[1]s
	tokn		%[2]d:%[3]s
	kind		%[4]v
	type		%[5]T
	valu		%[6]v

`, path, tokIdx, tok, curVal.Kind(), curVal.Interface(), string(buf))
}

// GetStr returns a string value from the config map.
func (c Config) GetStr(ctx context.Context, path string) string {
	return toString(c.Get(ctx, path))
}

// MarshalJSON implements a custom marshal routine for the Config type
// in order to omit the @parent@ field and prevent unnecessary data
// duplication in the marshaled output.
func (c Config) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	if _, err := w.Write([]byte{'{'}); err != nil {
		return nil, err
	}
	cx := 0
	cl := len(c)
	if _, ok := c[configScopeKey].(string); ok {
		cl--
	}
	if _, ok := c[configParentKey].(Config); ok {
		cl--
	}
	for k, v := range c {
		if k == configScopeKey || k == configParentKey {
			continue
		}
		if _, err := w.Write([]byte{'"'}); err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(k)); err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte{'"'}); err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte{':'}); err != nil {
			return nil, err
		}
		buf, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(buf); err != nil {
			return nil, err
		}
		if cx < cl-1 {
			if _, err := w.Write([]byte{','}); err != nil {
				return nil, err
			}
		}
		cx++
	}
	if _, err := w.Write([]byte{'}'}); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func getEnvVarName(path string) string {
	return fmt.Sprintf("LSX_%s",
		strings.Replace(strings.ToUpper(path), ".", "_", -1))
}

// isNillable returns two flags indicating whether or not the reflected
// value is of a nillable type and whether or not the value is nil.
// nillable types are listed at
// https://golang.org/pkg/reflect/#Value.IsNil
func isNillable(v reflect.Value) (bool, bool) {
	switch v.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Slice,
		reflect.Ptr:
		return true, v.IsNil()
	}
	return false, false
}

// derefValue loops until it returns a value that's neither a pointer
// or an interface
func derefValue(v reflect.Value) reflect.Value {
	for {
		k := v.Kind()

		// if the value is not a pointer or interface, or the value
		// is nil, return the value as it's not possible to dereference
		// the value any futher
		if (k != reflect.Ptr && k != reflect.Interface) || v.IsNil() {
			return v
		}

		v = v.Elem()
	}
}

func toString(v interface{}) string {
	return toStringWithOpts(v, true)
}

func toStringWithOpts(v interface{}, deep bool) string {
	// check to see if the value is a string-related type
	switch vt := v.(type) {
	case string:
		return vt
	case *string:
		return *vt
	case fmt.Stringer:
		return vt.String()
	case json.Marshaler:
		buf, err := vt.MarshalJSON()
		if err != nil {
			return err.Error()
		}
		return string(buf)
	case encoding.TextMarshaler:
		buf, err := vt.MarshalText()
		if err != nil {
			return err.Error()
		}
		return string(buf)
	}

	if !deep {
		return ""
	}

	// the value isn't a string related type. while it would be more
	// performant to stack endless more cases into the above switch
	// statement, for the sake of readability reflection is used to
	// determine the characteristics of the value
	vv := reflect.ValueOf(v)

	if _, isNil := isNillable(vv); isNil {
		return ""
	}

	vv = derefValue(vv)
	k := vv.Kind()

	// if the value is a string then return the string value
	if k == reflect.String {
		return vv.Interface().(string)
	}

	// if the value is a bool or numeric then return the value
	// formatted as a string with the %d format pattern
	if k >= reflect.Bool && k <= reflect.Complex128 {
		return fmt.Sprintf("%d", vv.Interface())
	}

	// if the value is an array or map then return the value
	// encoded as a JSON string
	if k == reflect.Array || k == reflect.Slice ||
		k == reflect.Map || k == reflect.Struct {
		buf, err := json.Marshal(vv.Interface())
		if err != nil {
			return err.Error()
		}
		return string(buf)
	}

	// if the value is a channel or function return the value
	// formatted as a string with the %T format pattern
	if k == reflect.Chan || k == reflect.Func {
		return fmt.Sprintf("%T", vv.Interface())
	}

	// if the value is an unsafe pointer return the value
	// formatted as a string with the %v format pattern
	if k == reflect.UnsafePointer {
		return fmt.Sprintf("unsafe.Pointer %v", vv.Interface())
	}

	return ""
}
