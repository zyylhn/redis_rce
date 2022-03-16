package utils

import (
	"github.com/gookit/color"
	"path/filepath"
	"strings"
)

var Red = color.FgRed.Render
var Yellow = color.FgLightYellow.Render
var LightGreen = color.FgLightGreen.Render
var LightCyan = color.FgLightCyan.Render

func GetBasePathFromPath(path string) string {
	if len(path)>1{
		if path[0:1]!="/"{
			path=strings.ReplaceAll(path,"\\","/")
		}
		path=strings.TrimSuffix(path,"/")
		path=filepath.Dir(path)
		path=strings.ReplaceAll(path,"\\","/")
		if path[0:1]!="/"{
			path=strings.ReplaceAll(path,"/","\\")
		}
	}
	return path
}

func GetFileNameFromPath(path string) string {
	if len(path)>1{
		if path[0:1]!="/"{
			path=strings.ReplaceAll(path,"\\","/")
		}
		path=strings.TrimSuffix(path,"/")
		path=filepath.Base(path)
		path=strings.ReplaceAll(path,"\\","/")
		if path[0:1]!="/"{
			path=strings.ReplaceAll(path,"/","\\")
		}
	}
	return path
}