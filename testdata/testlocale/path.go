package testlocale

import (
	"path"

	testdataloader "github.com/peteole/testdata-loader"
)

var (
	LocaleDir = path.Join(testdataloader.GetBasePath(), "testdata", "testlocale")
)
