
package kpsql

import (
	sql "database/sql"
	fmt "fmt"
)

type SqlTable interface{
	Create()(err error)
	Drop()(err error)
	SqlType()(*SqlType)
	HasTx()(bool)
	Begin()(error)
	Rollback()(error)
	Commit()(error)
	Insert(ins interface{})(n int64, err error)
	Delete(options ...sqloption)(n int64, err error)
	Update(ins interface{}, options ...sqloption)(n int64, err error)
	Select(options ...sqloption)(rows []interface{}, err error)
	SelectPrimary(ins interface{})(item interface{}, err error)
	Count(options ...sqloption)(n int64, err error)
}

type sqlTable struct{
	sqltype *SqlType
	name string
	sqldb SqlDatabase
	engine SqlEngine
	charset SqlCharset
	sqltx *sql.Tx
}

func NewSqlTable(sqldb SqlDatabase, name string, sqltype *SqlType, opts ...option)(tb *sqlTable){
	tb = &sqlTable{
		sqltype: sqltype,
		name: name,
		sqldb: sqldb,
		engine: ENGINE_INNODB,
		charset: CHAR_UTF8,
	}
	for _, o := range opts {
		o(tb)
	}
	return
}

func (tb *sqlTable)Create()(err error){
	var command string = "CREATE TABLE `" + tb.name + "` ("
	if len(tb.sqltype.tagList) > 0 {
		for _, tag := range tb.sqltype.tagList {
			command += "`" + tag.name + "` " + tag.word + " NOT NULL,"
		}
		if len(tb.sqltype.primaryKey) > 0 {
			command += "PRIMARY KEY ("
			for _, tag := range tb.sqltype.primaryKey {
				command += "`" + tag + "`,"
			}
			command = command[0:len(command) - 1] + "),"
		}
		command = command[0:len(command) - 1]
	}
	command += fmt.Sprintf(")ENGINE=%s DEFAULT CHARSET=%s", tb.engine.String(), tb.charset.String())

	_, err = tb.sqldb.DB().Exec(command)
	if err != nil { return }

	return nil
}

func (tb *sqlTable)Drop()(err error){
	var command string = "DROP TABLE `" + tb.name + "`"

	_, err = tb.sqldb.DB().Exec(command)
	if err != nil { return }

	return nil
}

func (tb *sqlTable)SqlType()(*SqlType){
	return tb.sqltype
}

func (tb *sqlTable)HasTx()(bool){
	return tb.sqltx != nil
}

func (tb *sqlTable)Begin()(err error){
	tb.Rollback()
	tb.sqltx, err = tb.sqldb.DB().Begin()
	return
}

func (tb *sqlTable)Rollback()(err error){
	if tb.sqltx == nil {
		return nil
	}
	err = tb.sqltx.Rollback()
	tb.sqltx = nil
	return
}

func (tb *sqlTable)Commit()(err error){
	if tb.sqltx == nil {
		return nil
	}
	err = tb.sqltx.Commit()
	tb.sqltx = nil
	return
}

func (tb *sqlTable)doWithTx(call func(*sql.Tx)(error))(err error){
	hastx := tb.sqltx != nil
	if !hastx {
		err = tb.Begin()
		if err != nil { return }
	}
	err = call(tb.sqltx)
	if err != nil {
		tb.Rollback()
		return
	}
	if !hastx {
		tb.Commit()
	}
	return nil
}

func (tb *sqlTable)Insert(ins interface{})(n int64, err error){
	err = tb.doWithTx(func(tx *sql.Tx)(err error){
		var stmt *sql.Stmt

		var command string = "INSERT INTO `" + tb.name + "` ("
		var values []interface{} = make([]interface{}, 0, len(tb.sqltype.fieldMap))
		revalue := getReValue(ins)
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
		n, err = res.LastInsertId()
		if err != nil { n = -1 }
		return nil
	})
	return
}

func (tb *sqlTable)Delete(options ...sqloption)(n int64, err error){
	err = tb.doWithTx(func(tx *sql.Tx)(err error){
		var stmt *sql.Stmt

		var opt = new(optionData).set(options...)
		var command string = "DELETE FROM `" + tb.name + "`"
		wcmd, values := opt.where.Format()
		command += wcmd + opt.makeTailCmd()

		stmt, err = tx.Prepare(command)
		if err != nil { return }
		defer stmt.Close()

		var res sql.Result
		res, err = stmt.Exec(values...)
		if err != nil { return }

		n, err = res.RowsAffected()
		if err != nil { n = -1 }
		return nil
	})
	return
}

func (tb *sqlTable)Update(ins interface{}, options ...sqloption)(n int64, err error){
	err = tb.doWithTx(func(tx *sql.Tx)(err error){
		var stmt *sql.Stmt

		var opt = new(optionData).set(options...)
		var command string = "UPDATE `" + tb.name + "` SET "
		var values []interface{}
		revalue := getReValue(ins)
		if opt.taglist == nil || len(opt.taglist) == 0 {
			values = make([]interface{}, 0, len(tb.sqltype.fieldMap))
			for _, tag := range tb.sqltype.fieldMap {
				if !tag.tagItem.primary_key {
					v := revalue.FieldByName(tag.Name).Interface()
					command += "`" + tag.tagItem.name + "` = ?,"
					values = append(values, sqlization(v))
				}
			}
		}else{
			values = make([]interface{}, 0, len(opt.taglist))
			for _, tag := range opt.taglist {
				v := revalue.FieldByName(tb.sqltype.fieldMap[tag].Name).Interface()
				command += "`" + tag + "` = ?,"
				values = append(values, sqlization(v))
			}
		}
		command = command[:len(command) - 1]
		if opt.where == nil && len(tb.sqltype.primaryKey) > 0 {
			opt.where = tb.sqltype.PriWhere(ins)
			opt.limit = []uint{1}
		}
		wcmd, where_values := opt.where.Format()
		command += wcmd + opt.makeTailCmd()
		values = append(values, where_values...)

		stmt, err = tx.Prepare(command)
		if err != nil { return }
		defer stmt.Close()

		var res sql.Result
		res, err = stmt.Exec(values...)
		if err != nil { return }

		n, err = res.RowsAffected()
		if err != nil { n = -1 }
		return nil
	})
	return
}

func (tb *sqlTable)Select(options ...sqloption)(items []interface{}, err error){
	err = tb.doWithTx(func(tx *sql.Tx)(err error){
		var (
			stmt *sql.Stmt
			rows *sql.Rows
		)

		var opt = new(optionData).set(options...)
		var scanRow []interface{} = make([]interface{}, 0, len(tb.sqltype.fieldMap))
		var command string = ""
		for _, tag := range tb.sqltype.tagList {
			command += "`" + tag.name + "`,"
			scanRow = append(scanRow, newByType(tb.sqltype.fieldMap[tag.name].Type))
		}
		command = "SELECT " + command[:len(command) - 1] + " FROM `" + tb.name + "`"
		wcmd, values := opt.where.Format()
		command += wcmd + opt.makeTailCmd()

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
		return nil
	})
	return
}

func (tb *sqlTable)SelectPrimary(ins interface{})(item interface{}, err error){
	items, err := tb.Select(tb.sqltype.PriWhereOpt(ins), OptLimit(1))
	if err != nil { return }
	if len(items) != 1 {
		return nil, nil
	}
	return items[0], nil
}

func (tb *sqlTable)Count(options ...sqloption)(n int64, err error){
	n = 0
	err = tb.doWithTx(func(tx *sql.Tx)(err error){
		var (
			stmt *sql.Stmt
			rows *sql.Rows
		)

		var opt = new(optionData).set(options...)
		var command string = "SELECT 1 FROM `" + tb.name + "`"
		wcmd, values := opt.where.Format()
		command += wcmd + opt.makeTailCmd()

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
		return nil
	})
	return
}
