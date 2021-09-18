
package kpsql

import (
	fmt "fmt"
)

type option func(tb *sqlTable)

type SqlEngine string

const (
	ENGINE_INNODB SqlEngine = "InnoDB"
)

func (e SqlEngine)String()(string){
	return (string)(e)
}

func OptionEngine(engine SqlEngine)(option){
	return func(tb *sqlTable){
		tb.engine = engine
	}
}

type SqlCharset string

const (
	CHAR_UTF8    SqlCharset = "utf8"
	CHAR_UTF8MB4 SqlCharset = "utf8mb4"
)

func (c SqlCharset)String()(string){
	return (string)(c)
}

func OptionCharset(charset SqlCharset)(option){
	return func(tb *sqlTable){
		tb.charset = charset
	}
}

type optionData struct{
	where WhereMap
	taglist []string
	limit []uint
	order string
}

func (opt *optionData)set(opts ...sqloption)(*optionData){
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func (opt *optionData)makeTailCmd()(cmd string){
	cmd = ""
	if opt.limit != nil {
		if len(opt.limit) > 1 {
			cmd += fmt.Sprintf(" LIMIT %d, %d", opt.limit[0], opt.limit[1])
		}else if len(opt.limit) > 0 {
			cmd += fmt.Sprintf(" LIMIT %d", opt.limit[0])
		}
	}
	if len(opt.order) != 0 {
		cmd += " ORDER BY " + opt.order
	}
	return
}

type sqloption func(data *optionData)

func OptWhere(where WhereMap)(sqloption){
	return func(data *optionData){
		data.where = where
	}
}

func OptWMap(arr ...interface{})(sqloption){
	where := MakeWMap(arr...)
	return func(data *optionData){
		data.where = where
	}
}

func OptWMapAnd(arr ...interface{})(sqloption){
	where := MakeWMapAnd(arr...)
	return func(data *optionData){
		data.where = where
	}
}

func OptWMapOr(arr ...interface{})(sqloption){
	where := MakeWMapOr(arr...)
	return func(data *optionData){
		data.where = where
	}
}

func OptWMapEq(arr ...interface{})(sqloption){
	where := MakeWMapEq(arr...)
	return func(data *optionData){
		data.where = where
	}
}

func OptWMapEqAnd(arr ...interface{})(sqloption){
	where := MakeWMapEqAnd(arr...)
	return func(data *optionData){
		data.where = where
	}
}

func OptWMapEqOr(arr ...interface{})(sqloption){
	where := MakeWMapEqOr(arr...)
	return func(data *optionData){
		data.where = where
	}
}

func OptLimit(limit ...uint)(sqloption){
	return func(data *optionData){
		data.limit = limit
	}
}

func OptTags(tags ...string)(sqloption){
	return func(data *optionData){
		data.taglist = tags
	}
}

func OptOrder(arr ...interface{})(sqloption){
	order := ""
	if len(arr) > 0 {
		var lo int8 = 0
		for i, o := range arr {
			switch o := o.(type) {
			case string:
				if lo == 1 {
					order += " ASC,"
				}
				order += "`" + o + "`"
				lo = 1
			case bool:
				if lo != 1 {
					panic("last arg must be string")
				}
				if o {
					order += " DESC,"
				}else{
					order += " ASC,"
				}
				lo = 2
			case int:
				if lo != 1 {
					panic("last arg must be string")
				}
				if o == 0 {
					order += " ASC,"
				}else{
					order += " DESC,"
				}
				lo = 2
			default:
				panic(fmt.Errorf("Unknown type: %T [%d]", o, i))
			}
		}
	}
	return func(data *optionData){
		data.order = order
	}
}
