package ai

import (
	"github.com/fatih/color"
)

// Colors for a nice ui
var (
	userColor	= color.New(color.FgCyan, color.Bold)
	aiColor		= color.New(color.FgGreen, color.Bold)
	systemColor = color.New(color.FgYellow)
	errorColor	= color.New(color.FgRed, color.Bold)
	debugColor	= color.New(color.FgMagenta)
	successColor= color.New(color.FgGreen)
)

