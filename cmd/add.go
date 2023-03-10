/*
Copyright © 2023 yixy <youzhilane01@gmail.com>
*/
package cmd

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/spf13/cobra"
	"github.com/yixy/tiny-photograph/internal"
	"github.com/yixy/tiny-photograph/internal/db"
	"github.com/yixy/tiny-photograph/internal/db/model"
	"github.com/yixy/tiny-photograph/internal/log"
	"go.uber.org/zap"
)

var taskId string
var rowNumImp = 0
var rowNumIgn = 0
var rowTotal = 0

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "specify a directory to add files",
	Long:  `specify a directory to add files by date and update metadata`,
	Run: func(cmd *cobra.Command, args []string) {

		et, err := exiftool.NewExiftool()
		if err != nil {
			log.Logger.Error("Error when initializing", zap.Error(err))
			return
		}
		defer et.Close()

		dir := args[0]
		files, err := os.ReadDir(dir)
		if err != nil {
			log.Logger.Error("Error when fetch the dir", zap.Error(err))
			return
		}

		taskId = time.Now().Format(time.RFC3339Nano)

		for _, file := range files {
			fileName := fmt.Sprintf("%s/%s", dir, file.Name())
			fileInfos := et.ExtractMetadata(fileName)

			//only return one file
			for _, fileInfo := range fileInfos {
				dealFile(fileInfo, file, fileName)
			}
		}

		//statics
		log.Logger.Info("statics", zap.Int("total", rowTotal), zap.Int("ignore", rowNumIgn), zap.String("taskId", taskId))
		fmt.Printf("taskId: %s\n", taskId)
		fmt.Printf("--------------------\n")
		tx, err := db.DB.Begin()
		if err != nil {
			fmt.Printf("statics error: %v \n", err)
			return
		}
		defer tx.Rollback()
		stmt, err := tx.PrepareContext(context.Background(), "select valid_flag,file_date,count(*) from file_obj_t where task_id=? group by valid_flag,file_date")
		if err != nil {
			fmt.Printf("statics error[PrepareContext]: %v \n", err)
			return
		}
		rows, err := stmt.QueryContext(context.Background(), taskId)
		if err != nil {
			fmt.Printf("statics error[QueryContext]: %v \n", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var flag, cnt int
			var d string
			err = rows.Scan(&flag, &d, &cnt)
			if err != nil {
				fmt.Printf("statics query rows error: %v", err)
				continue
			}
			log.Logger.Info("statics", zap.String("file_date", d), zap.Int("count", cnt), zap.Int("valid_flag", flag))
			rowNumImp += cnt
			fmt.Printf("%s|import|%10d files\n", d, cnt)
		}
		fmt.Printf("--------------------\n")
		fmt.Printf("import : %10d\n", rowNumImp)
		fmt.Printf("ignore: %10d\n", rowNumIgn)
		fmt.Printf("other failed: %10d\n", rowTotal-rowNumImp-rowNumIgn)
		fmt.Printf("total ignore: %10d\n", rowTotal)
		if err := rows.Err(); err != nil {
			fmt.Printf("statics rows.err: %v", err)
		}
	},
}

func dealFile(fileInfo exiftool.FileMetadata, file fs.DirEntry, fileName string) {
	ctx := context.Background()
	rowTotal++
	if fileInfo.Err != nil {
		log.Logger.Error(fmt.Sprintf("Error when reading file %v: %v\n", fileInfo.File, fileInfo.Err))
		return
	}

	var date interface{}
	var fileDate, fileTime, timeOrigin string
	ok := false
	const FileTypeExtension = "FileTypeExtension"
	const DateTimeOriginal = "DateTimeOriginal"
	const ModifyDate = "ModifyDate"
	const CreateDate = "CreateDate"
	const FileModifyDate = "FileModifyDate"

	//get fileType
	fileType, ok := fileInfo.Fields[FileTypeExtension].(string)
	if !ok {
		log.Logger.Error(fmt.Sprintf("%s fileType is not string", fileName))
		return
	}
	if !internal.IsTypeMatched(strings.ToLower(fileType)) {
		return
	}

	//get md5Sum
	src, err := os.Open(fileName)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("Error when open file %s", fileName), zap.Error(err))
		return
	}
	defer src.Close()
	h := md5.New()
	if _, err = io.Copy(h, src); err != nil {
		log.Logger.Error(fmt.Sprintf("Error when hash file %s", fileName), zap.Error(err))
		return
	}
	md5Sum := fmt.Sprintf("%x", h.Sum(nil))

	//check is exist
	fileObj := new(model.FileObj)
	err = fileObj.Get(ctx, md5Sum)
	if err == nil {
		log.Logger.Error(fmt.Sprintf("file exist: %v", fileName))
		rowNumIgn++
		return
	} else if err != sql.ErrNoRows {
		log.Logger.Error(fileName, zap.Error(err))
		return
	}

	now := time.Now()
	fileObj.Md5Sum = md5Sum
	fileObj.FileExtension = fileType
	fileObj.TaskId = taskId
	fileObj.ValidFlag = 1
	unixNano := now.UnixNano()
	fileObj.CreateTime = unixNano
	fileObj.UpdateTime = unixNano
	//Todo
	fileObj.TimeZone = "+08:00"

	//get fileTime
	if fileInfo.Fields[DateTimeOriginal] != nil {
		date = fileInfo.Fields[DateTimeOriginal]
		timeOrigin = DateTimeOriginal
	} else if fileInfo.Fields[ModifyDate] != nil {
		date = fileInfo.Fields[ModifyDate]
		timeOrigin = ModifyDate
	} else if fileInfo.Fields[CreateDate] != nil {
		date = fileInfo.Fields[CreateDate]
		timeOrigin = CreateDate
	} else if fileInfo.Fields[FileModifyDate] != nil {
		date = fileInfo.Fields[FileModifyDate]
		timeOrigin = FileModifyDate
	} else {
		date = time.Now().String()
		timeOrigin = "sysdate"
	}
	fileDate, ok = date.(string)
	if !ok {
		log.Logger.Error(fmt.Sprintf("%s fileDate is not string", fileName))
		return
	}
	fileDate = strings.ReplaceAll(fileDate, ":", "-")
	fileDate = strings.ReplaceAll(fileDate, " ", "_")
	fileTime = fileDate
	fileDate = fileDate[0:10]
	fileObj.TimeOrigin = timeOrigin
	fileObj.FileTime = fileTime
	fileObj.FileDate = fileDate

	//generate fileName
	newFileName := fmt.Sprintf("%s-%s.%s", fileTime, md5Sum, fileType)
	log.Logger.Info(fmt.Sprintf("%s [%v] %s\n", file.Name(), timeOrigin, newFileName))
	fmt.Printf("%s [%v] %s\n", file.Name(), timeOrigin, newFileName)
	fileObj.FileName = newFileName

	//copy file
	src.Seek(0, 0)

	tx, err := db.DB.Begin()
	if err != nil {
		log.Logger.Error("Error when start a sql tx ", zap.Error(err))
		return
	}
	defer tx.Rollback()
	err = fileObj.AddWithoutTx(ctx, tx)
	if err != nil {
		log.Logger.Error("Error when insert file meta data", zap.Error(err))
		return
	}
	targetDir := fmt.Sprintf("./db/%s", fileDate)
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("Error when mkdir %s", targetDir))
		return
	}
	targetPath2File := fmt.Sprintf("%s/%s", targetDir, newFileName)
	target, err := os.Create(targetPath2File)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("Error when open file %s/%s", targetDir, newFileName), zap.Error(err))
		return
	}
	defer target.Close()
	if _, err = io.Copy(target, src); err != nil {
		log.Logger.Error(fmt.Sprintf("Error when copy file %s", newFileName), zap.Error(err))
		return
	}
	tx.Commit()
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
