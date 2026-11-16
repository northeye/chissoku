package chissoku

const (
	programName = `chissoku`
	description = `A CO2 sensor reader`
	version     = "2.1.1" // x-release-please-version
)

// ProgramName returns the program name
func (c *Chissoku) ProgramName() string {
	return programName
}

// Version returns the program version
func (c *Chissoku) Version() string {
	return version
}

// Description returns the program description
func (c *Chissoku) Description() string {
	return description
}
