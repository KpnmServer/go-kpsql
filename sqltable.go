
package kpsql

import (
	sql "database/sql"
	fmt "fmt"
	reflect "reflect"
)

type WhereValue struct{
	Key   string
	Cond  string
	Value interface{}
	Next  string
}
type WhereMap []WhereValue

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

type SqlTable interface{
	Insert(ins interface{})(n int64, err error)
	Delete(wheremap WhereMap, limit ...uint)(n int64, err error)
	Update(ins interface{}, wheremap WhereMap, taglist []string, limit ...uint)(n int64, err error)
	Select(wheremap WhereMap, limit ...uint)(rows []interface{}, err error)
	Count(wheremap WhereMap, limit ...uint)(n int64, err error)
}

type sqlTable struct{
	sqltype *SqlType
	name string
	sqldb SqlDatabase
}

func NewSqlTable(sqldb SqlDatabase, name string, sqltype *SqlType)(tb SqlTable){
	return &sqlTable{
		sqltype: sqltype,
		name: name,
		sqldb: sqldb,
	}
}

func (tb *sqlTable)Insert(ins interface{})(n int64, err error){
	var(
		tx   *sql.Tx
		stmt *sql.Stmt
	)
	tx, err = tb.sqldb.DB().Begin()
	if err != nil { return }
	defer func(){ if tx != nil { tx.Rollback() } }()

	var command string = "INSERT INTO " + tb.name + " ("
	var values []interface{} = make([]interface{}, 0, len(tb.sqltype.fieldMap))
	revalue := reflect.ValueOf(ins).Elem()
	seats := ""
	for tag, field := range tb.sqltype.fieldMap {
		v := sqlizationReflect(revalue.FieldByName(field.Name))
		command += "`" + tag + "`,"
		seats += "?,"
		values = append(values, v)
	}
	command = command[:len(command) - 1] + ") VALUES (" + seats[:len(seats) - 1] + ")"

	stmt, err = tx.Prepare(command)
	if err != nil { return }
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(values...)
	if err != nil { return }
	tx.Commit(); tx = nil

	n, err = res.LastInsertId()
	if err != nil {
		n = -1
	}
	return n, nil
}

func (tb *sqlTable)Delete(argv WhereMap, limit ...uint)(n int64, err error){
	var(
		tx   *sql.Tx
		stmt *sql.Stmt
	)
	tx, err = tb.sqldb.DB().Begin()
	if err != nil { return }
	defer func(){ if tx != nil { tx.Rollback() } }()

	var command string =  "DELETE FROM " + tb.name
	wcmd, values := argv.Format()
	command += wcmd

	if len(limit) > 1 {
		command += fmt.Sprintf(" LIMIT %d, %d", limit[0], limit[1])
	}else if len(limit) > 0 {
		command += fmt.Sprintf(" LIMIT %d", limit[0])
	}

	stmt, err = tx.Prepare(command)
	if err != nil { return }
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(values...)
	if err != nil { return }
	tx.Commit(); tx = nil

	n, err = res.RowsAffected()
	if err != nil {
		n = -1
	}
	return n, nil
}

func (tb *sqlTable)Update(ins interface{}, argv WhereMap, taglist []string, limit ...uint)(n int64, err error){
	var(
		tx   *sql.Tx
		stmt *sql.Stmt
	)
	tx, err = tb.sqldb.DB().Begin()
	if err != nil { return }
	defer func(){ if tx != nil { tx.Rollback() } }()

	var command string = "UPDATE " + tb.name + " SET "
	var values []interface{}
	revalue := reflect.ValueOf(ins).Elem()
	if taglist == nil || len(taglist) == 0 {
		for tag, field := range tb.sqltype.fieldMap {
			if tag != tb.sqltype.primaryKey {
				v := revalue.FieldByName(field.Name).Interface()
				command += "`" + tag + "` = ?,"
				values = append(values, v)
			}
		}
	}else{
		for _, tag := range taglist {
			v := revalue.FieldByName(tb.sqltype.fieldMap[tag].Name).Interface()
			command += "`" + tag + "` = ?,"
			values = append(values, v)
		}
	}
	command = command[:len(command) - 1]
	if argv == nil && tb.sqltype.primaryKey != "" {
		argv = WhereMap{{tb.sqltype.primaryKey, "=",
			revalue.FieldByName(tb.sqltype.fieldMap[tb.sqltype.primaryKey].Name).Interface(), ""}}
	}
	wcmd, where_values := argv.Format()
	command += wcmd
	values = append(values, where_values...)

	if len(limit) > 1 {
		command += fmt.Sprintf(" LIMIT %d, %d", limit[0], limit[1])
	}else if len(limit) > 0 {
		command += fmt.Sprintf(" LIMIT %d", limit[0])
	}

	stmt, err = tx.Prepare(command)
	if err != nil { return }
	defer stmt.Close()

	var res sql.Result
	res, err = stmt.Exec(values...)
	if err != nil { return }
	tx.Commit(); tx = nil

	n, err = res.RowsAffected()
	if err != nil {
		n = -1
	}
	return n, nil
}

func (tb *sqlTable)Select(argv WhereMap, limit ...uint)(items []interface{}, err error){
	var(
		tx   *sql.Tx
		stmt *sql.Stmt
		rows *sql.Rows
	)
	tx, err = tb.sqldb.DB().Begin()
	if err != nil { return }
	defer func(){ if tx != nil { tx.Rollback() } }()

	var scanRow []interface{} = make([]interface{}, 0, len(tb.sqltype.fieldMap))
	var command string = ""
	for _, tag := range tb.sqltype.tagList {
		command += "`" + tag + "`,"
		scanRow = append(scanRow, newByType(tb.sqltype.fieldMap[tag].Type))
	}
	command = "SELECT " + command[:len(command) - 1] + " FROM " + tb.name
	wcmd, values := argv.Format()
	command += wcmd

	if len(limit) > 1 {
		command += fmt.Sprintf(" LIMIT %d, %d", limit[0], limit[1])
	}else if len(limit) > 0 {
		command += fmt.Sprintf(" LIMIT %d", limit[0])
	}

	stmt, err = tx.Prepare(command)
	if err != nil { return }
	defer stmt.Close()

	rows, err = stmt.Query(values...)
	if err != nil { return }
	defer rows.Close()

	items = make([]interface{}, 0)
	for rows.Next() {
		rowa := cloneValue(scanRow).([]interface{})
		err = rows.Scan(rowa...)
		if err != nil { return }
		items = append(items, tb.sqltype.newBy(rowa))
	}
	tx.Commit(); tx = nil

	return items, nil
}

func (tb *sqlTable)SelectPrimary(key interface{})(item interface{}, err error){
	items, err := tb.Select(WhereMap{{tb.sqltype.primaryKey, "=", key, ""}}, 1)
	if err != nil { return }
	if len(items) != 1 {
		return nil, nil
	}
	return items[0], nil
}

func (tb *sqlTable)Count(wheremap WhereMap, limit ...uint)(n int64, err error){
	n = 0
	var(
		tx   *sql.Tx
		stmt *sql.Stmt
		rows *sql.Rows
	)
	tx, err = tb.sqldb.DB().Begin()
	if err != nil { return }
	defer func(){ if tx != nil { tx.Rollback() } }()

	var command string = "SELECT 1 FROM " + tb.name
	wcmd, values := argv.Format()
	command += wcmd

	if len(limit) > 1 {
		command += fmt.Sprintf(" LIMIT %d, %d", limit[0], limit[1])
	}else if len(limit) > 0 {
		command += fmt.Sprintf(" LIMIT %d", limit[0])
	}

	stmt, err = tx.Prepare(command)
	if err != nil { return }
	defer stmt.Close()

	rows, err = stmt.Query(values...)
	if err != nil { return }
	defer rows.Close()

	a := new(int)
	for rows.Next() {
		err = rows.Scan(a)
		if err != nil { return }
		n += 1
	}
	tx.Commit(); tx = nil

	return n, nil
}
