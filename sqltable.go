
package kpsql

import (
	sql "database/sql"
	fmt "fmt"
)

type SqlTable interface{
	Name()(string)
	SqlType()(*SqlType)
	Create()(err error)
	Drop()(err error)
	Begin()(*SqlTx, error)
	Insert(ins interface{})(n int64, err error)
	Delete(options ...sqloption)(n int64, err error)
	Update(ins interface{}, options ...sqloption)(n int64, err error)
	Select(options ...sqloption)(rows []interface{}, err error)
	SelectPrimary(ins interface{}, options ...sqloption)(item interface{}, err error)
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

func (tb *sqlTable)Name()(string){
	return tb.name
}

func (tb *sqlTable)SqlType()(*SqlType){
	return tb.sqltype
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

func (tb *sqlTable)Begin()(tx *SqlTx, err error){
	var sqltx *sql.Tx
	sqltx, err = tb.sqldb.DB().Begin()
	if err != nil { return }
	return &SqlTx{
		table: tb,
		sqltx: sqltx,
	}, nil
}

func (tb *sqlTable)Insert(ins interface{})(n int64, err error){
	tx, err := tb.Begin()
	if err != nil { return }
	return tx.Insert(ins)
}

func (tb *sqlTable)Delete(options ...sqloption)(n int64, err error){
	tx, err := tb.Begin()
	if err != nil { return }
	return tx.Delete(options...)
}

func (tb *sqlTable)Update(ins interface{}, options ...sqloption)(n int64, err error){
	tx, err := tb.Begin()
	if err != nil { return }
	return tx.Update(ins, options...)
}

func (tb *sqlTable)Select(options ...sqloption)(items []interface{}, err error){
	tx, err := tb.Begin()
	if err != nil { return }
	return tx.Select(options...)
}

func (tb *sqlTable)SelectPrimary(ins interface{}, options ...sqloption)(item interface{}, err error){
	tx, err := tb.Begin()
	if err != nil { return }
	return tx.SelectPrimary(ins, options...)
}

func (tb *sqlTable)Count(options ...sqloption)(n int64, err error){
	tx, err := tb.Begin()
	if err != nil { return }
	return tx.Count(options...)
}
