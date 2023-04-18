package vm

import (
	"git.defalsify.org/vise/logging"
)

var (
	Logg logging.Logger = logging.NewVanilla().WithDomain("vm")
)
