package main

import (
	"fmt"
	"github.com/fatih/color"
)

var version = "dev"
var goVersion = "go"

var cyanBold = color.New(color.FgHiCyan).SprintFunc()
var whiteBold = color.New(color.Bold).SprintFunc()
var whiteDim = color.New(color.Faint).SprintFunc()
var redBold = color.New(color.Bold, color.FgHiRed).SprintFunc()
var greenBold = color.New(color.Bold, color.FgHiGreen).SprintFunc()
var yellowBold = color.New(color.Bold, color.FgHiYellow).SprintFunc()

func upkubeTextArt() string {
	return `
	█░█ █▀█ █▄▀ █░█ █▄▄ █▀▀
	█▄█ █▀▀ █░█ █▄█ █▄█ ██▄
	slim, purpose build 
	kubernetes deployment manager
`
}

func upkubeInfoMessage() string {
	return fmt.Sprintf("%s\tversion %s, build with Go %s\n", yellowBold(upkubeTextArt()), whiteBold(version), whiteBold(goVersion))
}
