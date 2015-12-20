package cqrs

import (
	"github.com/berkaroad/conf"
	"github.com/berkaroad/util"
	"log"
	"os"
)

var consoleLog = log.New(os.Stdout, "[ddd] ", log.LstdFlags)
var config = conf.LoadIniConfig(util.GetExecFilePath() + ".ini")
