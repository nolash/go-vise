package db

import (
	"git.defalsify.org/vise.git/logging"
)

var (
	Logg logging.Logger = logging.NewVanilla().WithDomain("db")
)
