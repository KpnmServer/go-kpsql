
package kpsql

import (
	fmt "fmt"
)

type WhereValue struct{
	Key   string
	Cond  string
	Value interface{}
	Next  string
}
type WhereMap []WhereValue

func MakeWMap(arr ...interface{})(wmap WhereMap){
	if len(arr) % 4 != 0 || len(arr) % 4 != 3 {
		panic("len(arr) % 4 != 0 || len(arr) % 4 != 3")
	}
	if len(arr) % 4 == 3 {
		arr = append(arr, "")
	}
	leng := len(arr)
	wmap = make(WhereMap, 0, leng)
	for i := 0; i < leng ;i += 4 {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: arr[i + 1].(string),
			Value: arr[i + 2],
			Next: arr[i + 3].(string),
		})
	}
	return
}

func MakeWMapAnd(arr ...interface{})(wmap WhereMap){
	if len(arr) % 3 != 0 {
		panic("len(arr) % 3 != 0")
	}
	leng := len(arr)
	wmap = make(WhereMap, 0, leng)
	for i := 0; i < leng ;i += 3 {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: arr[i + 1].(string),
			Value: arr[i + 2],
			Next: "AND",
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}

func MakeWMapOr(arr ...interface{})(wmap WhereMap){
	if len(arr) % 3 != 0 {
		panic("len(arr) % 3 != 0")
	}
	leng := len(arr)
	wmap = make(WhereMap, 0, leng)
	for i := 0; i < leng ;i += 3 {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: arr[i + 1].(string),
			Value: arr[i + 2],
			Next: "OR",
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}

func MakeWMapEq(arr ...interface{})(wmap WhereMap){
	if len(arr) % 3 != 0 || len(arr) % 3 != 2 {
		panic("len(arr) % 3 != 0 || len(arr) % 3 != 2")
	}
	if len(arr) % 3 == 2 {
		arr = append(arr, "")
	}
	leng := len(arr)
	wmap = make(WhereMap, 0, leng)
	for i := 0; i < leng ;i += 3 {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: "=",
			Value: arr[i + 1],
			Next: arr[i + 2].(string),
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}

func MakeWMapEqAnd(arr ...interface{})(wmap WhereMap){
	if len(arr) % 2 != 0 {
		panic("len(arr) % 2 != 0")
	}
	leng := len(arr)
	wmap = make(WhereMap, 0, leng)
	for i := 0; i < leng ;i += 2 {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: "=",
			Value: arr[i + 1],
			Next: "AND",
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}

func MakeWMapEqOr(arr ...interface{})(wmap WhereMap){
	if len(arr) % 2 != 0 {
		panic("len(arr) % 2 != 0")
	}
	leng := len(arr)
	wmap = make(WhereMap, 0, leng)
	for i := 0; i < leng ;i += 2 {
		wmap = append(wmap, WhereValue{
			Key: arr[i].(string),
			Cond: "=",
			Value: arr[i + 1],
			Next: "OR",
		})
	}
	if len(wmap) > 0 {
		wmap[len(wmap) - 1].Next = ""
	}
	return
}

func (wmap WhereMap)Format()(command string, values []interface{}){
	if wmap == nil || len(wmap) == 0 {
		return "", []interface{}{}
	}
	values = make([]interface{}, 0, len(wmap))
	command = " WHERE "
	for _, v := range wmap {
		command += fmt.Sprintf("`%s` %s ? %s", v.Key, v.Cond, v.Next)
		values = append(values, sqlization(v.Value))
	}
	command = command[:len(command) - 1 - len(wmap[len(wmap) - 1].Next)]
	return
}

