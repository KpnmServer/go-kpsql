
package kpsql

import (
	time "time"
	sql "database/sql"
)

const (
	DEFAULT_OPEN_CONN = 32
	DEFAULT_MAX_LIFE_TIME = 60 * time.Second
)

type SqlDatabase interface{
	DB()(*sql.DB)
	Close()(error)
	GetTable(name string, ins interface{})(SqlTable)
	GetTableBySqltype(name string, sqltype *SqlType)(SqlTable)
}

type sqlDatabase struct{
	dbins *sql.DB
}

func Open(name string, dbDSN string)(sqldb SqlDatabase, err error){
	db, err := sql.Open(name, dbDSN)
	sqldb = &sqlDatabase{dbins: db}
	if err != nil {
		return
	}

	db.SetMaxOpenConns(DEFAULT_OPEN_CONN)
	db.SetConnMaxLifetime(DEFAULT_MAX_LIFE_TIME)

	err = db.Ping()
	if err != nil {
		return
	}
	return sqldb, nil
}

func (sqldb *sqlDatabase)DB()(*sql.DB){
	return sqldb.dbins
}

func (sqldb *sqlDatabase)Close()(error){
	return sqldb.dbins.Close()
}

func (sqldb *sqlDatabase)GetTable(name string, ins interface{})(SqlTable){
	return sqldb.GetTableBySqltype(name, NewSqlType(ins))
}

func (sqldb *sqlDatabase)GetTableBySqltype(name string, sqltype *SqlType)(SqlTable){
	return NewSqlTable(sqldb, name, sqltype)
}


