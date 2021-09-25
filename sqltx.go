
package kpsql

import (
	sql "database/sql"
)

type SqlTx struct{
	table SqlTable
	sqltx *sql.Tx
}

func (tx *SqlTx)Rollback()(err error){
	if tx.sqltx == nil {
		return nil
	}
	err = tx.sqltx.Rollback()
	tx.sqltx = nil
	return
}

func (tx *SqlTx)Commit()(err error){
	if tx.sqltx == nil {
		return nil
	}
	err = tx.sqltx.Commit()
	tx.sqltx = nil
	return
}

func (tx *SqlTx)do(call func(*sql.Tx)(error))(err error){
	err = call(tx.sqltx)
	if err != nil {
		tx.Rollback()
		return
	}
	tx.Commit()
	return nil
}

func (tx *SqlTx)Insert(ins interface{})(n int64, err error){
	err = tx.do(func(sqltx *sql.Tx)(err error){
		var stmt *sql.Stmt

		var command string = "INSERT INTO `" + tx.table.Name() + "` ("
		var values []interface{} = make([]interface{}, 0, len(tx.table.SqlType().fieldMap))
		revalue := getReValue(ins)
		seats := ""
		for tag, field := range tx.table.SqlType().fieldMap {
			v := sqlizationReflect(revalue.FieldByName(field.Name))
			command += "`" + tag + "`,"
			seats += "?,"
			values = append(values, v)
		}
		command = command[:len(command) - 1] + ") VALUES (" + seats[:len(seats) - 1] + ")"

		stmt, err = sqltx.Prepare(command)
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

func (tx *SqlTx)Delete(options ...sqloption)(n int64, err error){
	err = tx.do(func(sqltx *sql.Tx)(err error){
		var stmt *sql.Stmt

		var opt = new(optionData).set(options...)
		var command string = "DELETE FROM `" + tx.table.Name() + "`"
		wcmd, values := opt.where.Format()
		command += wcmd + opt.makeTailCmd()

		stmt, err = sqltx.Prepare(command)
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

func (tx *SqlTx)Update(ins interface{}, options ...sqloption)(n int64, err error){
	err = tx.do(func(sqltx *sql.Tx)(err error){
		var stmt *sql.Stmt

		var opt = new(optionData).set(options...)
		var command string = "UPDATE `" + tx.table.Name() + "` SET "
		var values []interface{}
		revalue := getReValue(ins)
		if opt.taglist == nil || len(opt.taglist) == 0 {
			values = make([]interface{}, 0, len(tx.table.SqlType().fieldMap))
			for _, tag := range tx.table.SqlType().fieldMap {
				if !tag.tagItem.primary_key {
					v := revalue.FieldByName(tag.Name).Interface()
					command += "`" + tag.tagItem.name + "` = ?,"
					values = append(values, sqlization(v))
				}
			}
		}else{
			values = make([]interface{}, 0, len(opt.taglist))
			for _, tag := range opt.taglist {
				v := revalue.FieldByName(tx.table.SqlType().fieldMap[tag].Name).Interface()
				command += "`" + tag + "` = ?,"
				values = append(values, sqlization(v))
			}
		}
		command = command[:len(command) - 1]
		if opt.where == nil && len(tx.table.SqlType().primaryKey) > 0 {
			opt.where = tx.table.SqlType().PriWhere(ins)
			opt.limit = []uint{1}
		}
		wcmd, where_values := opt.where.Format()
		command += wcmd + opt.makeTailCmd()
		values = append(values, where_values...)

		stmt, err = sqltx.Prepare(command)
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

func (tx *SqlTx)Select(options ...sqloption)(items []interface{}, err error){
	err = tx.do(func(sqltx *sql.Tx)(err error){
		var (
			stmt *sql.Stmt
			rows *sql.Rows
		)

		var opt = new(optionData).set(options...)
		var scanRow []interface{} = make([]interface{}, 0, len(tx.table.SqlType().fieldMap))
		var command string = ""
		if opt.taglist != nil && len(opt.taglist) > 0 {
			for _, tag := range opt.taglist {
				tg, ok := tx.table.SqlType().fieldMap[tag]
				if ok {
					command += "`" + tag + "`,"
					scanRow = append(scanRow, newByType(tg.Type))
				}
			}
		}else{
			for _, tag := range tx.table.SqlType().tagList {
				command += "`" + tag.name + "`,"
				scanRow = append(scanRow, newByType(tx.table.SqlType().fieldMap[tag.name].Type))
			}
		}
		command = "SELECT " + command[:len(command) - 1] + " FROM `" + tx.table.Name() + "`"
		wcmd, values := opt.where.Format()
		command += wcmd + opt.makeTailCmd()

		stmt, err = sqltx.Prepare(command)
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
			items = append(items, tx.table.SqlType().newBy(rowa))
		}
		return nil
	})
	return
}

func (tx *SqlTx)SelectPrimary(ins interface{}, options ...sqloption)(item interface{}, err error){
	options = append(options, tx.table.SqlType().PriWhereOpt(ins), OptLimit(1))
	items, err := tx.table.Select(options...)
	if err != nil { return }
	if len(items) != 1 {
		return nil, nil
	}
	return items[0], nil
}

func (tx *SqlTx)Count(options ...sqloption)(n int64, err error){
	n = 0
	err = tx.do(func(sqltx *sql.Tx)(err error){
		var (
			stmt *sql.Stmt
			rows *sql.Rows
		)

		var opt = new(optionData).set(options...)
		var command string = "SELECT 1 FROM `" + tx.table.Name() + "`"
		wcmd, values := opt.where.Format()
		command += wcmd + opt.makeTailCmd()

		stmt, err = sqltx.Prepare(command)
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

