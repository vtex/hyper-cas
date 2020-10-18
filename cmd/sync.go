package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vtex/hyper-cas/synchronizer"
)

var syncLabel string
var syncURL string
var syncJson bool
var syncRetries int

func folderExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && info.Mode().IsDir()
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync a folder into a distribution in hyper-cas",
	Long:  `Sync will synchronize all files in a given folder into hyper-cas`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			panic("There must be only a single argument specifying path to sync.")
		}
		var err error
		folder := args[0]
		if !filepath.IsAbs(folder) {
			folder, err = filepath.Abs(folder)
			if err != nil {
				panic(err)
			}
		}
		if !folderExists(folder) {
			panic(fmt.Sprintf("Folder %s does not exist!", folder))
		}
		s := synchronizer.NewSync(folder, syncURL)
		var result map[string]interface{}
		retries := 0
		for i := 0; i <= syncRetries; i++ {
			result, err = s.Run(syncLabel)
			if err == nil {
				break
			}
			retries += 1
		}
		if result == nil {
			panic(fmt.Errorf("Failed to synchronize folder: %v", err))
		}
		result["retries"] = retries
		if syncJson {
			res, err := json.Marshal(result)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(res))
		} else {
			printResult(result)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVarP(&syncLabel, "label", "l", "", "Label to apply to this new distribution")
	syncCmd.Flags().StringVarP(&syncURL, "api-url", "u", "http://localhost:2485/", "Hyper-CAS API URL")
	syncCmd.Flags().BoolVarP(&syncJson, "json", "j", false, "Whether to output JSON serialization")
	syncCmd.Flags().IntVarP(&syncRetries, "retries", "r", 3, "Number of times to retry synchronizing")
}

func printResult(result map[string]interface{}) {
	fmt.Printf("Completed synchronizing with %d retries.\n", result["retries"].(int))
	for _, file := range result["files"].([]map[string]interface{}) {
		isUpToDate := file["upToDate"].(bool)
		path := file["path"].(string)
		if isUpToDate {
			fmt.Printf("* %s - Already up-to-date.\n", path)
		} else {
			fmt.Printf("* %s - Updated (hash: %s).\n", path, file["hash"].(string))
		}
	}

	distro := result["distro"].(map[string]interface{})
	fmt.Printf("* Distro %s is up-to-date.\n", distro["hash"].(string))

	label := result["label"].(map[string]interface{})
	labelName := label["label"].(string)
	if labelName != "" {
		fmt.Printf("* Updated label %s => %s.\n", labelName, label["hash"].(string))
	}
}
