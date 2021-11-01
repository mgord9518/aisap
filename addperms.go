package aisap

import (
//	"bytes"
//	"io/ioutil"
//	"os"
//	"os/user"
//	"os/exec"
//	"strings"
//	"strconv"
//	"errors"
	"path/filepath"
//
//	//appimage "github.com/probonopd/go-appimage/src/goappimage"
//	helpers  "github.com/mgord9518/aisap/helpers"
//	profiles "github.com/mgord9518/aisap/profiles"
//	ini	  "gopkg.in/ini.v1"
//	xdg	  "github.com/adrg/xdg"
)

// Add extra FS permissions before parsing the desktop entry
func AddFilePerms(s []string) {
	for i, val := range(s) {
		end := val[len(val)-3:]

		if end != ":ro" && end != ":rw" {
			s[i] = s[i]+":ro"
		}
	}

	preFilePerms = s
}

func AddSharePerms(s []string) {
	preSharePerms = s
}

func AddSocketPerms(s []string) {
	preSocketPerms = s
}

func AddDevicePerms(slice []string) {
	for i, _ := range slice {
		device := filepath.Clean(slice[i])
		if len(device) < 5 || device [0:5] != "/dev/" {
			device = filepath.Clean("/dev/"+device)
		}
		slice[i] = device
	}
	preDevicePerms = slice
}
