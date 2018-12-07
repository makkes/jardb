package jsondb

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/makkes/jardb/db"
)

type JSONDB struct {
	Entries  map[string][]string
	JarFiles map[string]JarFile
}

type JarFile struct {
	Name string
	Date time.Time
}

func (jf JarFile) hash() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(jf.Name)))
}

func (db *JSONDB) Find(pattern string) <-chan string {
	pattern = strings.Replace(pattern, "/", ".", -1)
	pattern = regexp.MustCompile("\\.class$").ReplaceAllString(pattern, "")
	out := make(chan string)
	go func() {
		for class, jarFiles := range db.Entries {
			m, err := regexp.MatchString(pattern, class)
			if m && err == nil {
				for _, jarFile := range jarFiles {
					out <- fmt.Sprintf("%s: %s", class, db.JarFiles[jarFile].Name)
				}
			}
		}
		close(out)
	}()

	return out
}

func (db *JSONDB) IndexFolders(root string) {
	entries, err := ioutil.ReadDir(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %s\n", root, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			db.IndexFolders(path.Join(root, entry.Name()))
		} else {
			if strings.HasSuffix(entry.Name(), ".jar") {
				db.indexFile(path.Join(root, entry.Name()))
			}
		}
	}
}

func (db *JSONDB) indexFile(file string) {
	r, err := zip.OpenReader(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n", file, err)
		return
	}
	defer r.Close()

	osf, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n", file, err)
		return
	}
	defer osf.Close()

	var modTime time.Time
	stat, err := osf.Stat()
	if err != nil {
		modTime = time.Now()
	} else {
		modTime = stat.ModTime()
	}

	jarFile := JarFile{
		Name: file,
		Date: modTime,
	}
	indexedJarFile := db.JarFiles[jarFile.hash()]
	if (indexedJarFile != JarFile{}) && !indexedJarFile.Date.Before(modTime) {
		return
	}

	fmt.Printf("Indexing %s\n", file)

	db.JarFiles[jarFile.hash()] = jarFile

	for _, classFile := range r.File {
		if !classFile.FileInfo().IsDir() && strings.HasSuffix(classFile.Name, ".class") && !strings.HasSuffix(classFile.Name, "package-info.class") {
			className := strings.Replace(strings.TrimSuffix(classFile.Name, ".class"), "/", ".", -1)
			entry := db.Entries[className]

			if entry == nil {
				db.Entries[className] = []string{
					jarFile.hash(),
				}
			} else {
				entryExists := false
				for _, jarFilePointer := range db.Entries[className] {
					if jarFilePointer == jarFile.hash() {
						entryExists = true
						break
					}
				}
				if !entryExists {
					db.Entries[className] = append(db.Entries[className], jarFile.hash())
				}
			}
		}
	}
}

func NewJSONDB() *JSONDB {
	in, err := os.OpenFile("/home/max/.jardb.json", os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("%#v\n", err)
		panic(err)
	}
	defer in.Close()

	bytes, err := ioutil.ReadAll(in)
	if err != nil {
		panic(err)
	}

	var db JSONDB
	if len(bytes) == 0 {
		db.Entries = make(map[string][]string)
		db.JarFiles = make(map[string]JarFile)
		return &db
	}
	err = json.Unmarshal(bytes, &db)
	if err != nil {
		panic(err)
	}

	return &db
}

func (db *JSONDB) Close() error {
	bytes, err := json.Marshal(db)
	if err != nil {
		return err
	}
	out, err := os.OpenFile("/home/max/.jardb.json", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	out.Write(bytes)
	return nil
}

func (d *JSONDB) Stats() db.Stats {
	return db.Stats{
		ClassCount: len(d.Entries),
		JarCount:   len(d.JarFiles),
	}
}
