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
	"os"
	"path/filepath"
	"strings"
	"sync"
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
var cache = NewTypeCache()
var rowNumMatch = 0
var rowNumImp = 0
var rowNumIgn = 0
var rowTotal = 0
var impDir *string
var label *string
var fileMonth *bool

type TypeCache struct {
	cache map[string]bool
	sync.Mutex
}

func NewTypeCache() *TypeCache {
	return &TypeCache{
		cache: make(map[string]bool),
	}
}

func (c *TypeCache) Add(t string) {
	c.Lock()
	defer c.Unlock()
	c.cache[t] = true
}

func (c *TypeCache) String() (result string) {
	c.Lock()
	defer c.Unlock()
	for k, _ := range c.cache {
		result = fmt.Sprintf("%s, %s", result,k)
	}
	return result
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "specify a directory to add files",
	Long:  `specify a directory to add files by date and update metadata`,
	Run: func(cmd *cobra.Command, args []string) {

		buf := make([]byte, 128*1000)
		et, err := exiftool.NewExiftool(exiftool.Buffer(buf, 64*1000))
		if err != nil {
			log.Logger.Error("Error when initializing", zap.Error(err))
			return
		}
		defer et.Close()

		taskId = time.Now().Format(time.RFC3339Nano)

		// Use filepath.Walk to traverse the directory
		err = filepath.Walk(*impDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Check if current path refers to a file
			if info.IsDir() {
				return nil
			}
			baseName := info.Name()
			//fileName := fmt.Sprintf("%s/%s", path, baseName)
			fileName := path
			fileInfos := et.ExtractMetadata(fileName)

			//only return one file
			for _, fileInfo := range fileInfos {
				dealFile(fileInfo, baseName, fileName)
			}
			return nil
		})
		if err != nil {
			log.Logger.Error("Error when travels dir", zap.Error(err))
			//return
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
		stmt, err := tx.PrepareContext(context.Background(), "select valid_flag,file_month,count(*) from file_obj_t where task_id=? group by valid_flag,file_month")
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
			log.Logger.Info("statics", zap.String("file_month", d), zap.Int("count", cnt), zap.Int("valid_flag", flag))
			rowNumImp += cnt
			fmt.Printf("%s|import|%10d files\n", d, cnt)
		}
		fmt.Printf("--------------------\n")
		fmt.Printf("import : %10d\n", rowNumImp)
		fmt.Printf("ignore: %10d\n", rowNumIgn)
		fmt.Printf("not matched: %10d\n", rowNumMatch)
		fmt.Printf("other failed: %10d\n", rowTotal-rowNumImp-rowNumIgn-rowNumMatch)
		fmt.Printf("total: %10d\n", rowTotal)
		fmt.Printf("not matched type: %s \n",cache.String())
		if err := rows.Err(); err != nil {
			fmt.Printf("statics rows.err: %v", err)
		}
	},
}

func dealFile(fileInfo exiftool.FileMetadata, baseName string, fileName string) {
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
	const FileType = "FileType"
	const DateTimeOriginal = "DateTimeOriginal"
	const ModifyDate = "ModifyDate"
	const CreateDate = "CreateDate"
	const FileModifyDate = "FileModifyDate"

	//get fileType
	var fileTypeExtension string
	fileTypeExtensionInfo := fileInfo.Fields[FileTypeExtension]
	if fileTypeExtensionInfo == nil {
		log.Logger.Warn(fmt.Sprintf("%s fileTypeExtension is nil", fileName))
	}else {
		var ok bool
		fileTypeExtension, ok = fileTypeExtensionInfo.(string)
		if !ok {
			log.Logger.Warn(fmt.Sprintf("%s fileTypeExtension is not string", fileName))
		}
	}

	fileTypeInfo := fileInfo.Fields[FileType]
	if fileTypeInfo == nil {
		log.Logger.Error(fmt.Sprintf("%s fileType is nil", fileName))
		return
	}
	fileType, ok := fileTypeInfo.(string)
	if !ok {
		log.Logger.Error(fmt.Sprintf("%s fileType is not string", fileName))
		return
	}
	if !internal.IsTypeMatched(strings.ToLower(fileType)) {
		log.Logger.Warn(fmt.Sprintf("%s fileType is not matched: %s", fileName, fileType))
		rowNumMatch++

		// 缓存类型
		cache.Add(fileType)
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
		log.Logger.Info(fmt.Sprintf("file exist: %v", fileName))
		rowNumIgn++
		return
	} else if err != sql.ErrNoRows {
		log.Logger.Error(fileName, zap.Error(err))
		return
	}

	now := time.Now()
	fileObj.Md5Sum = md5Sum
	fileObj.FileType = fileType
	fileObj.FileExtension = fileTypeExtension
	fileObj.TaskId = taskId
	fileObj.ValidFlag = 1
	unixNano := now.UnixNano()
	fileObj.CreateTime = unixNano
	fileObj.UpdateTime = unixNano
	//TODO
	fileObj.TimeZone = "+08:00"

	//set task label
	if *label != "" {
		fileObj.Label = *label
	}

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
	if len(fileDate)<19{
		log.Logger.Error(fmt.Sprintf("%s fileDate format error, less than 19: %s", fileName,fileDate))
		return
	}
	//ignore zone in fileDate
	fileDate = fileDate[0:19]
	fileDate = strings.ReplaceAll(fileDate, ":", "-")
	fileDate = strings.ReplaceAll(fileDate, " ", "_")
	fileDate = strings.ReplaceAll(fileDate, ".", "-")
	fileTime = fileDate
	fileDate = fileDate[0:10]
	fileObj.TimeOrigin = timeOrigin
	fileObj.FileTime = fileTime
	fileObj.FileDate = fileDate
	fileObj.FileMonth = fileDate[0:7]
	if *fileMonth{
		// 获取上层目录名
		dirName := filepath.Dir(fileName)
		// 获取上层目录名的最后一个元素
		_, dir := filepath.Split(dirName)
		fileObj.FileMonth=dir
	}

	//generate fileName
	newFileName := fmt.Sprintf("%s-%s.%s", fileTime, md5Sum, fileType)
	log.Logger.Info(fmt.Sprintf("%s [%v] %s\n", baseName, timeOrigin, newFileName))
	fmt.Printf("%s [%v] %s\n", baseName, timeOrigin, newFileName)
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
	targetDir := fmt.Sprintf("./db/%s", fileObj.FileMonth)
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
	impDir = addCmd.Flags().StringP("dir", "d", "", "specify import directory")
	label = addCmd.Flags().StringP("label", "l", "", "specify task label")
	fileMonth = addCmd.Flags().BoolP("month", "m", false, "specify file month by dir")
}
