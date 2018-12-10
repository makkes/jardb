package boltdb

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/makkes/jardb/db"
)

type BoltDB struct {
	*bolt.DB
}

type JarFile struct {
	Name string
	Date time.Time
}

func (jf JarFile) hash() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(jf.Name)))
}

func (d *BoltDB) Find(pattern string) <-chan string {
	pattern = strings.Replace(pattern, "/", ".", -1)
	pattern = regexp.MustCompile("\\.class$").ReplaceAllString(pattern, "")
	out := make(chan string)
	go func() {
		defer close(out)
		err := d.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("classes"))
			return bucket.ForEach(func(className, v []byte) error {
				m, err := regexp.MatchString(pattern, string(className))
				if m && err == nil {
					var hashes []string
					buf := bytes.NewBuffer(v)
					dec := gob.NewDecoder(buf)
					dec.Decode(&hashes)
					for _, hash := range hashes {
						jarFile, err := d.getJarFile(hash)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error fetching jar file from index: %s\n", err)
							continue
						}
						out <- fmt.Sprintf("%s: %s", string(className), jarFile.Name)
					}
				}
				return nil
			})
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying index: %s\n", err)
			return
		}
	}()

	return out
}

func (d *BoltDB) IndexFolders(root string) {
	entries, err := ioutil.ReadDir(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %s\n", root, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			d.IndexFolders(path.Join(root, entry.Name()))
		} else {
			if strings.HasSuffix(entry.Name(), ".jar") {
				d.indexFile(path.Join(root, entry.Name()))
			}
		}
	}
}

func (d *BoltDB) getJarFile(hash string) (JarFile, error) {
	var res JarFile
	err := d.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("jarfiles"))
		if bucket == nil {
			return nil
		}
		v := bucket.Get([]byte(hash))
		buf := bytes.NewBuffer(v)
		dec := gob.NewDecoder(buf)
		dec.Decode(&res)
		return nil
	})
	return res, err
}

func (d *BoltDB) putJarFile(jarFile JarFile) error {
	err := d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("jarfiles"))
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		enc.Encode(jarFile)
		return bucket.Put([]byte(jarFile.hash()), buf.Bytes())
	})
	return err
}

func (d *BoltDB) getEntry(className string) ([]string, error) {
	var res []string
	err := d.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("classes"))
		if bucket == nil {
			return nil
		}
		v := bucket.Get([]byte(className))
		buf := bytes.NewBuffer(v)
		dec := gob.NewDecoder(buf)
		dec.Decode(&res)
		return nil
	})
	return res, err
}

func (d *BoltDB) putEntry(className string, entry []string) error {
	err := d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("classes"))
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		enc.Encode(entry)
		return bucket.Put([]byte(className), buf.Bytes())
	})
	return err
}

func (d *BoltDB) indexFile(file string) {
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

	indexedJarFile, err := d.getJarFile(jarFile.hash())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching %s from index: %s\n", jarFile.Name, err)
		return
	}
	if !indexedJarFile.Date.Before(modTime) {
		return
	}

	fmt.Printf("Indexing %s\n", file)

	if d.putJarFile(jarFile) != nil {
		fmt.Fprintf(os.Stderr, "Error writing to index: %s\n", err)
		return
	}

	for _, classFile := range r.File {
		if !classFile.FileInfo().IsDir() && strings.HasSuffix(classFile.Name, ".class") && !strings.HasSuffix(classFile.Name, "package-info.class") {
			className := strings.Replace(strings.TrimSuffix(classFile.Name, ".class"), "/", ".", -1)
			entry, err := d.getEntry(className)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching entry from %s: %s\n", jarFile.Name, err)
				continue
			}

			if entry == nil {
				err := d.putEntry(className, []string{
					jarFile.hash(),
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error storing entry in index: %s\n", err)
					continue
				}
			} else {
				entryExists := false
				existingEntry, err := d.getEntry(className)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error storing entry in index: %s\n", err)
					continue
				}
				for _, jarFilePointer := range existingEntry {
					if jarFilePointer == jarFile.hash() {
						entryExists = true
						break
					}
				}
				if !entryExists {
					d.putEntry(className, append(entry, jarFile.hash()))
				}
			}
		}
	}
}

func NewBoltDB() *BoltDB {
	d, err := bolt.Open(path.Join(os.Getenv("HOME"), ".jardb.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %s\n", err)
		return nil
	}
	return &BoltDB{d}
}

func (d *BoltDB) Stats() db.Stats {
	res := db.Stats{}
	d.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("classes"))
		if bucket != nil {
			res.ClassCount = bucket.Stats().KeyN
		}
		bucket = tx.Bucket([]byte("jarfiles"))
		if bucket != nil {
			res.JarCount = bucket.Stats().KeyN
		}
		return nil
	})
	return res
}
