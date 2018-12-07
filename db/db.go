package db

type Stats struct {
	ClassCount int
	JarCount   int
}

type DB interface {
	Find(pattern string) <-chan string
	IndexFolders(root string)
	Close() error
	Stats() Stats
}
