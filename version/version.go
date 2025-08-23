package version

import (
	"time"
)

// This is used within the gorelease flags
var (
	version = "Dev"
	Commit  = "None"
	Date    = time.Now()
)
