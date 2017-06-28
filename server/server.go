package server

// Holds the connection driver.
type Connection struct {
	Driver Driver
}

// Connection parameters.
type Params struct {
	Host       string
	Port       int
	Username   string
	Password   string
	Privatekey string
	Path       string
	Maxclients int
}

// Initialise a connection with a driver.
func Manage(driver Driver) *Connection {
	return &Connection{Driver: driver}
}

// Create a directory.
func (c *Connection) MkDir(path string) error {
	if err := c.Driver.MkDir(path); err != nil {
		return err
	}

	return nil
}

// Upload a file.
func (c *Connection) Upload(path, destination string) error {
	if err := c.Driver.Upload(path, destination); err != nil {
		return err
	}

	return nil
}

// Read a file contents.
func (c *Connection) Read(path string) (string, error) {
	contents, err := c.Driver.Read(path)
	if err != nil {
		return "", err
	}

	return contents, nil
}

// Delete a file.
func (c *Connection) Delete(path string) error {
	if err := c.Driver.Delete(path); err != nil {
		return err
	}

	return nil
}

// Execute a command on the server.
func (c *Connection) Exec(command string) (string, error) {
	return c.Driver.Exec(command)
}

// Close connection.
func (c *Connection) Close() {
	c.Driver.Close()
}
