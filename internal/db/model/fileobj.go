package model

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yixy/tiny-photograph/internal/db"
)

type FileObj struct {
	Md5Sum        string `json:"md5_hex"`
	FileName      string `json:"file_name"`
	FileExtension string `json:"file_extension"`
	FileTime      string `json:"file_time"`
	TimeZone      string `json:"time_zone"`
	TimeOrigin    string `json:"time_origin"`
	Label         string `json:"label"`
	TaskId        string `json:"task_id"`
	CreateTime    int64  `json:"create_time"`
	UpdateTime    int64  `json:"update_time"`
	ValidFlag     int    `json:"valid_flag"`
}

var objAddWithoutTx = func(ctx context.Context, tx *sql.Tx, args ...interface{}) error {
	obj, ok := args[0].(*FileObj)
	if !ok {
		return errors.New("objAdd:fileObj parse err")
	}

	stmt, err := tx.PrepareContext(ctx, "insert into file_obj_t(md5_hex,file_name,file_extension,file_time,time_zone,time_origin,label,task_id,create_time,update_time,valid_flag)values(?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, obj.Md5Sum, obj.FileName, obj.FileExtension, obj.FileTime, obj.TimeZone, obj.TimeOrigin, obj.Label, obj.TaskId, obj.CreateTime, obj.UpdateTime, obj.ValidFlag)
	if err != nil {
		return err
	}
	return nil
}

var objAdd = db.ExecuteSql(objAddWithoutTx)

var objGet = db.ExecuteSql(func(ctx context.Context, tx *sql.Tx, args ...interface{}) error {
	obj, ok := args[0].(*FileObj)
	if !ok {
		return errors.New("objGet:fileObj parse err")
	}

	stmt, err := tx.PrepareContext(ctx, "select md5_hex,file_name,file_extension,file_time,time_zone,time_origin,label,task_id,create_time,update_time,valid_flag from file_obj_t where md5_hex=? and valid_flag='1'")
	if err != nil {
		return err
	}

	err = stmt.QueryRowContext(ctx, obj.Md5Sum).Scan(&obj.Md5Sum, &obj.FileName, &obj.FileExtension, &obj.FileTime, &obj.TimeZone, &obj.TimeOrigin, &obj.Label, &obj.TaskId, &obj.CreateTime, &obj.UpdateTime, &obj.ValidFlag)
	if err != nil {
		return err
	}
	return nil
})

func (obj *FileObj) Add(ctx context.Context) error {
	return objAdd(ctx, obj)
}
func (obj *FileObj) AddWithoutTx(ctx context.Context, tx *sql.Tx) error {
	return objAddWithoutTx(ctx, tx, obj)
}
func (obj *FileObj) Get(ctx context.Context, md5Sum string) error {
	obj.Md5Sum = md5Sum
	return objGet(ctx, obj)
}
