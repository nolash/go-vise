package mem

import (
	"git.defalsify.org/vise.git/logging"
)

var (
	logg logging.Logger = logging.NewVanilla().WithDomain("memdb")
)