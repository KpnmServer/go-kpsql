
package kpsql

import (
	strings "strings"
	reflect "reflect"
)

type tagItem struct{
	name string
	word string
	foreign_key [2]string
	primary_key bool
}

type fieldValue struct{
	Name string
	Type reflect.Type
	tagItem *tagItem
}

type SqlType struct{
	instance interface{}
	retype reflect.Type
	primaryKey []string
	tagList []tagItem
	fieldMap map[string]fieldValue
	priwmap WhereMap
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
		primaryKey: make([]string, 0, 1),
		tagList: make([]tagItem, 0, nf),
		fieldMap: make(map[string]fieldValue),
		priwmap: nil,
	}
	for i := 0; i < nf ;i++ {
		field := retype.Field(i)
		tag := field.Tag
		sqltag, ok := tag.Lookup("sql")
		if !ok {
			continue
		}
		var isprimary bool = false
		if ispri, ok := tag.Lookup("sql_primary"); ok && ispri == "true" {
			sqltype.primaryKey = append(sqltype.primaryKey, sqltag)
			isprimary = true
		}
		sqlword := tag.Get("sqlword")
		sfk := tag.Get("sql_foreign_key")
		foreign_key := [2]string{"", ""}
		if ind := strings.IndexByte(sfk, ':'); ind >= 0 {
			foreign_key[0], foreign_key[1] = sfk[0:ind], sfk[ind + 1:len(sfk)]
		}
		tagitem := tagItem{
			name: sqltag, word: sqlword, foreign_key: foreign_key, primary_key: isprimary,
		}
		sqltype.tagList = append(sqltype.tagList, tagitem)
		sqltype.fieldMap[tagitem.name] = fieldValue{
			Name: field.Name,
			Type: field.Type,
			tagItem: &tagitem,
		}
	}
	return
}

func (sqltype *SqlType)PriWhere(ins interface{})(WhereMap){
	if sqltype.priwmap == nil {
		arr := make([]interface{}, len(sqltype.primaryKey) * 2)
		revalue := getReValue(ins)
		for i, k := range sqltype.primaryKey {
			arr[i] = k
			arr[i + 1] = revalue.FieldByName(sqltype.fieldMap[k].Name).Interface()
		}
		sqltype.priwmap = MakeWMapEqAnd(arr...)
	}
	return sqltype.priwmap
}

func (sqltype *SqlType)PriWhereOpt(ins interface{})(sqloption){
	return OptWhere(sqltype.PriWhere(ins))
}

func (sqltype *SqlType)newBy(row []interface{})(obj interface{}){
	revalue := cloneReflectValue(reflect.ValueOf(sqltype.instance))
	for i, tag := range sqltype.tagList {
		fie := sqltype.fieldMap[tag.name]
		field := revalue.FieldByName(fie.Name)
		setReflectValue(field, row[i])
	}
	return revalue.Addr().Interface()
}
