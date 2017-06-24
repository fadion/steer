package server

// Holds the connection driver.
type Connection struct {
	driver Driver
}

// Connection parameters.
type Params struct {
	Host       string
	Port       int
	Username   string
	Password   string
	Privatekey string
	Path       string
}

// Initialise a connection with a driver.
func Manage(driver Driver) *Connection {
	return &Connection{driver: driver}
}

// Create a directory.
func (c *Connection) MkDir(path string) error {
	if err := c.driver.MkDir(path); err != nil {
		return err
	}

	return nil
}

// Upload a file.
func (c *Connection) Upload(path, destination string) error {
	if err := c.driver.Upload(path, destination); err != nil {
		return err
	}

	return nil
}

// Read a file contents.
func (c *Connection) Read(path string) (string, error) {
	contents, err := c.driver.Read(path)
	if err != nil {
		return "", err
	}

	return contents, nil
}

// Delete a file.
func (c *Connection) Delete(path string) error {
	if err := c.driver.Delete(path); err != nil {
		return err
	}

	return nil
}

// Close connection.
func (c *Connection) Close() {
	c.driver.Close()
}
