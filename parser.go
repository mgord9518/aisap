package aisap

import (
    "path/filepath"
    //"strings"

//    xdg     "github.com/adrg/xdg"
    helpers "github.com/mgord9518/aisap/helpers"
)

func parseDevicePerms(devices string, permLevel int) []string {
    s := helpers.DesktopSlice(devices)

    for i, _ := range s {
        device := filepath.Clean(s[i])
        if len(device) < 5 && device[0:5] == "/dev/" { // In case "../" was used, test against the expanded directory and make sure it's in /dev
			device = filepath.Clean(device[6:])
        }
        s[i] = device
		print(s[i])
    }

    return s
}

// Parses requested sockets into bwrap flags
func parseSocketPerms(sockets string, permLevel int) []string {
    return helpers.DesktopSlice(sockets)
}


// Parses requested sockets into bwrap flags
func parseSharePerms(share string, permLevel int) []string {
    return helpers.DesktopSlice(share)
}

// Parses requested file and directories into bwrap flags
//func parseFilePerms(directories string, permLevel int) []string) {
//    var genericDir string
//    dirPerms := make(map[string]string)
//
//    // Map out the XDG directories
//    var xdgDirs = map[string]string {
//        "xdg-home":        xdg.Home,
//        "xdg-desktop":     xdg.UserDirs.Desktop,
//        "xdg-download":    xdg.UserDirs.Download,
//        "xdg-documents":   xdg.UserDirs.Documents,
//        "xdg-music":       xdg.UserDirs.Music,
//        "xdg-pictures":    xdg.UserDirs.Pictures,
//        "xdg-videos":      xdg.UserDirs.Videos,
//        "xdg-templates":   xdg.UserDirs.Templates,
//        "xdg-publicshare": xdg.UserDirs.PublicShare,
//        "xdg-config":      xdg.ConfigHome,
//        "xdg-cache":       xdg.CacheHome,
//        "xdg-data":        xdg.DataHome,
//    }
//
//    // Anonymize directories by giving them generic names in case the user has
//    // changed the location of their XDG-dirs
//    var xdgGeneric = map[string]string {
//        "xdg-home":        homed,
//        "xdg-desktop":     homed+"/Desktop",
//        "xdg-download":    homed+"/Downloads",
//        "xdg-documents":   homed+"/Documents",
//        "xdg-music":       homed+"/Music",
//        "xdg-pictures":    homed+"/Pictures",
//        "xdg-videos":      homed+"/Videos",
//        "xdg-templates":   homed+"/Templaates",
//        "xdg-publicshare": homed+"/Share",
//        "xdg-config":      homed+"/.config",
//        "xdg-cache":       homed+"/.cache",
//        "xdg-data":        homed+"/.local/share",
//    }
//
//    s := helpers.DesktopSlice(directories)
//    for i, _ := range s {
//        str := s[i]
//
//        // If neither "ro" or "rw" provided, assume "ro"
//        l := len(str) //
//        if l <= 3 || str[l-3:] != ":ro" && str[l-3:] != ":rw" {
//           str = str+":ro"
//        }
//
//        // Replace the xdg-* strings with the corresponding directories on the user's machine
//        for key, val := range xdgDirs {
//
//            // If length of key bigger than requested directory or not equal to it continue because there is no reason to look at it further
//            if len(key) > len(str) || key != str[:len(key)] {
//                continue
//            }
//
//            // If the last byte of the requested path shortened to key length is a '/' or ':' we know it's the parent dir, so resolve it using the xdgDirs map
//            c := str[len(key)]          // The final byte of the key (used for splitting)
//            r := str[len(key):] // Every string after that byte
//            if c == byte('/') || c == byte(':') {
//                genericDir = xdgGeneric[key] + strings.Split(r, ":")[0]
//                s[i] = strings.Replace(str, key, val, 1)
//                break
//            } else {
//                genericDir = strings.Split(str, ":")[0]
//            }
//        }
//
//        // Separate the directory from "ro"/"rw" options
////        dir := strings.Split(s[i], ":")[0] // The directory at hand
//
//        dirPerms[s[i]] = genericDir
//    }
//
//    return dirPerms
//}
func parseFilePerms(directories string, permLevel int) []string {
    //dirPerms := make(map[string]string)

    // Map out the XDG directories

    s := helpers.DesktopSlice(directories)
    for i, _ := range s {
        //s[i]

        // If neither "ro" or "rw" provided, assume "ro"
        l := len(s[i]) //
        if l <= 3 || s[i][l-3:] != ":ro" && s[i][l-3:] != ":rw" {
           s[i] = s[i]+":ro"
        }

        // Separate the directory from "ro"/"rw" options
//        dir := strings.Split(s[i], ":")[0] // The directory at hand

        //dirPerms[s[i]] = genericDir
    }

    return s
}
