package server

// Server connection driver.
type Driver interface {
	MkDir(path string) error
	Upload(path string, destination string) error
	Read(path string) (string, error)
	Delete(path string) error
	Close()
}
