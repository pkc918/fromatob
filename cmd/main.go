package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	ValidFolder = map[string]bool{
		".git":       true,
		".idea":      true,
		".vscode":    true,
		".gitignore": true,
	}

	FileList   []string
	FolderList []string

	SourceDir  string
	TargetDirs []string
)

var RootCmd = &cobra.Command{
	Use:   "froma2b",
	Short: "Move a to b",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(args, SourceDir, TargetDirs)

		if _, err := os.Stat(SourceDir); err != nil {
			return fmt.Errorf("source direactory does not exist: %s", SourceDir)
		}

		for _, target := range TargetDirs {
			if _, err := os.Stat(target); os.IsNotExist(err) {
				return fmt.Errorf("target directory does not exist: %s", target)
			}
		}

		FilePathWalk()
		FromA2B()
		return nil
	},
}

func init() {
	initCobra()
}

func main() {
	Execute()
}

func initCobra() {
	RootCmd.Flags().StringVarP(&SourceDir, "from", "f", "", "Source directory(required)")
	_ = RootCmd.MarkFlagRequired("from")

	RootCmd.Flags().StringSliceVarP(&TargetDirs, "to", "t", []string{}, "Target directories")
	_ = RootCmd.MarkFlagRequired("to")
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func FilePathWalk() {
	_ = filepath.Walk(SourceDir, GetAllFile)
	for _, toPath := range TargetDirs {
		_ = filepath.Walk(toPath, PostAllFolder(toPath))
	}
}

func PostAllFolder(toPath string) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if IsSkipFilename(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			relativePath, _ := filepath.Rel(toPath, path)
			isTargetFolder := strings.Contains(relativePath, "/")
			if isTargetFolder {
				FolderList = append(FolderList, path)
			}
		}
		return nil
	}
}

// GetAllFile 获取需要移动的所有文件
func GetAllFile(path string, info fs.FileInfo, err error) error {
	if err != nil {
		fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
		return err
	}
	if IsSkipFilename(info.Name()) {
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}
	if !info.IsDir() {
		FileList = append(FileList, path)
	}
	//fmt.Println(path)
	return nil
}

// IsSkipFilename 是否是需要跳过的文件或者文件夹
func IsSkipFilename(filename string) bool {
	_, exists := ValidFolder[filename]
	return exists
}

func FromA2B() {
	filesLen := len(FileList)
	for i, folder := range FolderList {
		ii := i % (filesLen - 1)
		_ = CopyFile(FileList[ii], folder)
	}
}

func CopyFile(source, destination string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", "copy", source, destination)
	case "darwin", "linux":
		cmd = exec.Command("cp", source, destination)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to copy file: %s, output: %s", err, string(output))
	}
	return nil
}
