
package kpsql

import (
	time "time"
	hex "encoding/hex"
	reflect "reflect"
)

func newByType(ttype reflect.Type)(interface{}){
	if ttype.Kind() == reflect.Ptr {
		ttype = ttype.Elem()
	}
	switch ttype.Kind() {
	case reflect.Bool:
		return new(bool)
	case reflect.Int:
		return new(int)
	case reflect.Int8:
		return new(int8)
	case reflect.Int16:
		return new(int16)
	case reflect.Int32:
		return new(int32)
	case reflect.Int64:
		return new(int64)
	case reflect.Uint:
		return new(uint)
	case reflect.Uint8:
		return new(uint8)
	case reflect.Uint16:
		return new(uint16)
	case reflect.Uint32:
		return new(uint32)
	case reflect.Uint64:
		return new(uint64)
	case reflect.Float32:
		return new(float32)
	case reflect.Float64:
		return new(float64)
	case reflect.String:
		return new(string)
	case reflect.Array:
		return new(string)
	default:
		switch ttype.Name() {
		case "time.Time":
			return new(time.Time)
		}
	}
	panic("Unknow type " + ttype.Name())
	return reflect.New(ttype)
}

func setReflectValue(rvalue reflect.Value, value interface{}){
	rtype := rvalue.Type()
	switch rtype.Kind() {
	case reflect.Array:
		hexstr := value.(string)
		bytes, err := hex.DecodeString(hexstr)
		if err == nil {
			rvalue.SetBytes(bytes)
		}
	default:
		rvalue.Set(reflect.ValueOf(value).Elem())
	}
}

func getReflectValue(rvalue reflect.Value)(interface{}){
	rtype := rvalue.Type()
	switch rtype.Kind() {
	case reflect.Array:
		bytes := rvalue.Bytes()
		hexstr := hex.EncodeToString(bytes)
		return hexstr
	default:
		return rvalue.Interface()
	}
}

func cloneReflectValue(basevalue reflect.Value)(revalue reflect.Value){
	point_count := 0
	for basevalue.Kind() == reflect.Ptr {
		basevalue = basevalue.Elem()
		point_count++
	}
	
	switch basevalue.Kind() {
	case reflect.Slice:
		bl := basevalue.Len()
		revalue = reflect.MakeSlice(basevalue.Type(), basevalue.Len(), basevalue.Cap())
		for i := 0; i < bl ;i++ {
			revalue.Index(i).Set(cloneReflectValue(basevalue.Index(i)))
		}
	case reflect.Struct:
		revalue = reflect.New(basevalue.Type()).Elem()
		nf := revalue.NumField()
		for i := 0; i < nf ;i++ {
			field := revalue.Field(i)
			fieldtype := revalue.Type().Field(i)
			if 'A' < fieldtype.Name[0] && fieldtype.Name[0] < 'Z' {
				field.Set(cloneReflectValue(basevalue.Field(i)))
			}
		}
	default:
		revalue = reflect.New(basevalue.Type()).Elem()
		revalue.Set(basevalue)
	}
	for ;point_count > 0 ;point_count-- {
		revalue = revalue.Addr()
	}
	return
}

func cloneValue(base interface{})(re interface{}){
	return cloneReflectValue(reflect.ValueOf(base)).Interface()
}