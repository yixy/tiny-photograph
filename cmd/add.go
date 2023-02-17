/*
Copyright © 2023 yixy <youzhilane01@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/spf13/cobra"
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
			panic(err)
		}
		for _, file := range files {

			fileInfos := et.ExtractMetadata(fmt.Sprintf("%s/%s", dir, file.Name()))

			for _, fileInfo := range fileInfos {
				if fileInfo.Err != nil {
					fmt.Printf("Error concerning %v: %v\n", fileInfo.File, fileInfo.Err)
					continue
				}

				var date interface{}
				var fileDate, dateMark string
				ok := false
				const FileTypeExtension = "FileTypeExtension"
				const DateTimeOriginal = "DateTimeOriginal"
				const ModifyDate = "ModifyDate"
				const CreateDate = "CreateDate"
				const FileModifyDate = "FileModifyDate"
				fileType, ok := fileInfo.Fields[FileTypeExtension].(string)
				if !ok {
					fmt.Println("fileType is not string")
					continue
				}
				if strings.ToLower(fileType) == "jpg" {
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
						fmt.Println("fileDate is not string")
						continue
					}
					fileDate = fileDate[0:10]
					fileDate = strings.ReplaceAll(fileDate, ":", "-")

					fmt.Printf("%s [%v] %s \n", file.Name(), dateMark, fileDate)
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
