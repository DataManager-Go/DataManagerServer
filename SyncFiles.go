package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/JojiiOfficial/gaw"
)

func syncFiles(dryRun, noconfirm bool) error {
	// List locally available files
	files, err := ioutil.ReadDir(config.Server.PathConfig.FileStore)
	if err != nil {
		return err
	}

	// Get Local files+ID
	rows, err := db.Table("files").Where("deleted_at is NULL").Select("id, local_name").Rows()
	if err != nil {
		return err
	}

	var untrackedInFS []int64  // Database file rows without local file
	var untrackedInDB []string // Local files which don't have any reference in database
	var wg sync.WaitGroup

	wg.Add(2)
	cerr := make(chan error, 2)

	go func() {
		untrackedInFS, err = searchLocalFS(rows, files)
		cerr <- err
		wg.Done()
	}()

	go func() {
		untrackedInDB, err = searchDB(files)
		cerr <- err
		wg.Done()
	}()

	wg.Wait()
	close(cerr)

	for i := range cerr {
		if i != nil {
			return i
		}
	}

	if len(untrackedInDB) == 0 && len(untrackedInFS) == 0 {
		fmt.Println("Nothing to do")
		return nil
	}

	fmt.Println()
	fmt.Printf("Found %d files missing in local filesystem\n", len(untrackedInFS))
	fmt.Printf("Found %d files untracked in database\n", len(untrackedInDB))

	if dryRun {
		return nil
	}

	if !noconfirm {
		if y, _ := gaw.ConfirmInput("\nDelete all untracked files? [y/n/a]> ", bufio.NewReader(os.Stdin)); !y {
			return nil
		}
	}

	fmt.Println("Deleting...")

	return nil
}

// Returns a slice of FileIDs which have no local file
func searchLocalFS(rows *sql.Rows, files []os.FileInfo) ([]int64, error) {
	var untrackedInFS []int64

	var fileName string
	var id int64
	for rows.Next() {
		if err := rows.Scan(&id, &fileName); err != nil {
			return nil, err
		}

		if !hasLocalFile(files, fileName) {
			untrackedInFS = append(untrackedInFS, id)
		}
	}

	return untrackedInFS, nil
}

// Returns a slice of local files which don't have a reference in DB
func searchDB(localFiles []os.FileInfo) ([]string, error) {
	var untrackedInDB []string
	var i, j, end int

	for {
		if i >= len(localFiles) {
			break
		}
		end = i + 1000000

		if end > len(localFiles) {
			end = len(localFiles)
		}

		var res []string
		fmt.Println(i, end)
		if err := db.Table("files").Where("local_name NOT IN (?) AND deleted_at IS NULL", fifoToStr(localFiles[i:end])).Select("local_name").Find(&res).Error; err != nil {
			return nil, err
		}

		untrackedInDB = append(untrackedInDB, res...)
		i = end + 1
		j++
	}

	return untrackedInDB, nil
}

func fifoToStr(f []os.FileInfo) []string {
	s := make([]string, len(f))
	for i := range f {
		s[i] = f[i].Name()
	}
	return s
}

func hasLocalFile(localFiles []os.FileInfo, name string) bool {
	for i := range localFiles {
		if localFiles[i].Name() == name {
			return true
		}
	}

	return false
}
