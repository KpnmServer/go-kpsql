
package kpsql

import (
	reflect "reflect"
)

type SqlType struct{
	instance interface{}
	retype reflect.Type
	tagList []string
	fieldMap map[string]reflect.StructField
}

func NewSqlType(ins interface{})(sqltype *SqlType){
	retype := reflect.TypeOf(ins)
	if retype.Kind() == reflect.Ptr {
		retype = retype.Elem()
		ins = reflect.ValueOf(ins).Elem().Interface()
	}
	nf := retype.NumField()
	sqltype = &SqlType{
		instance: ins,
		retype: retype,
		tagList: make([]string, 0, nf),
		fieldMap: make(map[string]reflect.StructField),
	}
	for i := 0; i < nf ;i++ {
		field := retype.Field(i)
		tag := field.Tag
		sqltag, ok := tag.Lookup("sql")
		if !ok {
			continue
		}
		sqltype.tagList = append(sqltype.tagList, sqltag)
		sqltype.fieldMap[sqltag] = field
	}
	return
}

func (sqltype *SqlType)newBy(row []interface{})(obj interface{}){
	revalue := cloneReflectValue(reflect.ValueOf(sqltype.instance))
	for i, tag := range sqltype.tagList {
		fie := sqltype.fieldMap[tag]
		field := revalue.FieldByName(fie.Name)
		field.Set(reflect.ValueOf(row[i]).Elem())
	}
	return revalue.Interface()
}
