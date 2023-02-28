package model

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yixy/tiny-photograph/internal/db"
)

type FileObj struct {
	Md5Sum          string `json:"md5_hex"`
	FileName        string `json:"file_name"`
	FileExtension   string `json:"file_extension"`
	FileTime        string `json:"file_time"`
	TimeZone        string `json:"time_zone"`
	TimeOrigin      string `json:"time_origin"`
	Label           string `json:"label"`
	TaskId          string `json:"task_id"`
	CreateTime      int    `json:"create_time"`
	UpdateTime      int    `json:"update_time"`
	CreateLocalTime int    `json:"create_local_time"`
	UpdateLocalTime int    `json:"update_local_time"`
	ValidFlag       int    `json:"valid_flag"`
}

var objAdd = db.ExecuteSql(func(ctx context.Context, tx *sql.Tx, args ...interface{}) error {
	obj, ok := args[0].(*FileObj)
	if !ok {
		return errors.New("objAdd:fileObj parse err")
	}

	stmt, err := tx.PrepareContext(ctx, "insert into file_obj_t(md5_hex,file_name,file_extension,file_time,time_zone,time_origin,label,task_id,create_time,update_time,create_local_time,update_local_time,valid_flag)values(?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, obj.Md5Sum, obj.FileName, obj.FileExtension, obj.FileTime, obj.TimeZone, obj.TimeOrigin, obj.Label, obj.TaskId, obj.CreateTime, obj.UpdateTime, obj.CreateLocalTime, obj.UpdateLocalTime, obj.ValidFlag)
	if err != nil {
		return err
	}
	return nil
})

//var objGet = db.ExecuteSql(func(ctx context.Context, tx *sql.Tx, args ...interface{}) error {
//	_, err := tx.ExecContext(ctx, "insert into file_obj_t(md5_hex,file_name,file_extension,file_time,time_zone,time_origin,label,task_id,create_time,update_time,create_local_time,update_local_time,valid_flag)values(?,?,?,?,?,?,?,?,?,?,?,?,?)")
//	return err
//})

func (obj *FileObj) Add(ctx context.Context, args ...interface{}) error {
	return objAdd(ctx, obj)
}
