package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/yixy/golang-util/path"
	"github.com/yixy/tiny-photograph/common/env"
)

const dbFile = env.AppName + ".db"

// The returned DB is safe for concurrent use by multiple goroutines
// and maintains its own pool of idle connections. Thus, the Open
// function should be called just once. It is rarely necessary to
// close a DB.
var DB *sql.DB

// var SerializableTx *sql.TxOptions = &sql.TxOptions{
// 	Isolation: sql.LevelSerializable,
// 	ReadOnly:  false}

func init() {
	var err error
	DB, err = getConnection()
	if err != nil {
		panic(errors.WithMessage(err, "db getConnection error"))
	}
	err = ExecuteSqlFile("conf/sql/ddl.sql")
	if err != nil {
		panic(err)
	}
}

func getConnection() (*sql.DB, error) {
	workdir, err := path.GetProgramPath()
	if err != nil {
		panic(err)
	}
	return sql.Open("sqlite3", fmt.Sprintf("%s/%s?_txlock=exclusive", workdir, dbFile))
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

func ExecuteSql(fn func(context.Context, *sql.Tx, ...interface{}) error) func(context.Context, ...interface{}) error {
	return func(ctx context.Context, args ...interface{}) error {
		tx, err := DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		//execute sql
		err = fn(ctx, tx, args...)
		if err != nil {
			return err
		}

		tx.Commit()
		return nil
	}
}

func ExecuteSqlFile(filePath string) error {
	ddl, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//execute sql
	_, err = tx.Exec(string(ddl))
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}
