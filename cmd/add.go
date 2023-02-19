/*
Copyright © 2023 yixy <youzhilane01@gmail.com>
*/
package cmd

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/spf13/cobra"
	"github.com/yixy/tiny-photograph/internal"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "specify a directory to add files",
	Long:  `specify a directory to add files by date and update metadata`,
	Run: func(cmd *cobra.Command, args []string) {

		et, err := exiftool.NewExiftool()
		if err != nil {
			fmt.Printf("Error when intializing: %v\n", err)
			return
		}
		defer et.Close()

		dir := args[0]
		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("Error when fetch the dir : %v\n", err)
			return
		}

		for _, file := range files {
			fileName := fmt.Sprintf("%s/%s", dir, file.Name())
			fileInfos := et.ExtractMetadata(fileName)
			for _, fileInfo := range fileInfos {
				if fileInfo.Err != nil {
					fmt.Printf("Error when reading file %v: %v\n", fileInfo.File, fileInfo.Err)
					continue
				}

				var date interface{}
				var fileDate, fileTime, dateMark string
				ok := false
				const FileTypeExtension = "FileTypeExtension"
				const DateTimeOriginal = "DateTimeOriginal"
				const ModifyDate = "ModifyDate"
				const CreateDate = "CreateDate"
				const FileModifyDate = "FileModifyDate"
				fileType, ok := fileInfo.Fields[FileTypeExtension].(string)
				if !ok {
					fmt.Printf("%s fileType is not string", fileName)
					continue
				}
				if internal.IsTypeMatched(strings.ToLower(fileType)) {
					if fileInfo.Fields[DateTimeOriginal] != nil {
						date = fileInfo.Fields[DateTimeOriginal]
						dateMark = DateTimeOriginal
					} else if fileInfo.Fields[ModifyDate] != nil {
						date = fileInfo.Fields[ModifyDate]
						dateMark = ModifyDate
					} else if fileInfo.Fields[CreateDate] != nil {
						date = fileInfo.Fields[CreateDate]
						dateMark = CreateDate
					} else if fileInfo.Fields[FileModifyDate] != nil {
						date = fileInfo.Fields[FileModifyDate]
						dateMark = FileModifyDate
					} else {
						date = time.Now().String()
						dateMark = "sysdate"
					}
					fileDate, ok = date.(string)
					if !ok {
						fmt.Printf("%s fileDate is not string", fileName)
						continue
					}
					fileDate = strings.ReplaceAll(fileDate, ":", "-")
					fileDate = strings.ReplaceAll(fileDate, " ", "_")
					fileTime = fileDate
					fileDate = fileDate[0:10]

					//open file handle
					f, err := os.Open(fileName)
					if err != nil {
						fmt.Printf("Error when open file %s: %v\n", fileName, err)
						continue
					}
					defer f.Close()

					h := md5.New()
					if _, err = io.Copy(h, f); err != nil {
						fmt.Printf("Error when hash file %s: %v\n", fileName, err)
						continue
					}
					md5Sum := h.Sum(nil)
					newFileName := fmt.Sprintf("%s-%x.%s", fileTime, md5Sum, fileType)
					fmt.Printf("%s [%v] %s\n", file.Name(), dateMark, newFileName)

					src, err := os.Open(fileName)
					if err != nil {
						fmt.Printf("Error when open file %s: %v\n", fileName, err)
						continue
					}
					defer src.Close()

					targetDir := fmt.Sprintf("./db/%s", fileDate)
					err = os.MkdirAll(targetDir, 0755)
					if err != nil {
						fmt.Printf("Error when mkdir %s", targetDir)
						return
					}
					target, err := os.Create(fmt.Sprintf("%s/%s", targetDir, newFileName))
					if err != nil {
						fmt.Printf("Error when open file %s/%s: %v\n", targetDir, newFileName, err)
						continue
					}
					defer target.Close()
					if _, err = io.Copy(target, src); err != nil {
						fmt.Printf("Error when copy file %s: %v\n", newFileName, err)
						return
					}
				}
			}
		}
	},
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
