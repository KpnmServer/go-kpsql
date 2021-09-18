
package kpsql

import (
	time "time"
	hex "encoding/hex"
	reflect "reflect"
)

func sqlization(value interface{})(interface{}){
	return sqlizationReflect(reflect.ValueOf(value))
}

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
		case "time.Time", "Time":
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
		if rtype.Elem().Kind() == reflect.Uint8 {
			hexstr := *(value.(*string))
			bytes, err := hex.DecodeString(hexstr)
			if err == nil {
				bytesToByteArr(bytes, rvalue)
			}
			return
		}
	}
	rvalue.Set(reflect.ValueOf(value).Elem())
}

func sqlizationReflect(rvalue reflect.Value)(interface{}){
	rtype := rvalue.Type()
	switch rtype.Kind() {
	case reflect.Array:
		if rtype.Elem().Kind() == reflect.Uint8 {
			bytes := byteArrToBytes(rvalue)
			hexstr := hex.EncodeToString(bytes)
			return hexstr
		}
	}
	return rvalue.Interface()
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

func byteArrToBytes(rvalue reflect.Value)(bytes []byte){
	bytes = make([]byte, rvalue.Len())
	for i, _ := range bytes {
		bytes[i] = (byte)(rvalue.Index(i).Uint())
	}
	return
}

func bytesToByteArr(bytes []byte, bytearr reflect.Value){
	for i, b := range bytes {
		bytearr.Index(i).SetUint((uint64)(b))
	}
}

func getReValue(ins interface{})(revalue reflect.Value){
	revalue = reflect.ValueOf(ins)
	if revalue.Type().Kind() == reflect.Ptr {
		revalue = revalue.Elem()
	}
	return revalue
}

