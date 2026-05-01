package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func Create(dataPath string, backupDir string) (string, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("2006-01-02T15-04-05")
	filename := filepath.Join(backupDir, fmt.Sprintf("escrowd-backup-%s.tar.gz", timestamp))

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	paths := []string{dataPath, dataPath + "-audit"}
	for _, path := range paths {
		err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			header.Name = filePath
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if !info.IsDir() {
				f, err := os.Open(filePath)
				if err != nil {
					return err
				}
				defer f.Close()
				_, err = io.Copy(tw, f)
				return err
			}
			return nil
		})
		if err != nil {
			return "", err
		}
	}

	return filename, nil
}

func StartScheduled(dataPath string, backupDir string, interval time.Duration) {
	go func() {
		fmt.Printf("backup scheduler started — running every %s\n", interval)
		for {
			time.Sleep(interval)
			filename, err := Create(dataPath, backupDir)
			if err != nil {
				fmt.Println("backup failed:", err)
				continue
			}
			fmt.Printf("backup created: %s\n", filename)
			cleanOldBackups(backupDir, 7)
		}
	}()
}

func cleanOldBackups(backupDir string, keepCount int) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	var backups []string
	for _, e := range entries {
		if !e.IsDir() {
			backups = append(backups, filepath.Join(backupDir, e.Name()))
		}
	}

	if len(backups) <= keepCount {
		return
	}

	for _, old := range backups[:len(backups)-keepCount] {
		os.Remove(old)
		fmt.Println("removed old backup:", old)
	}
}
