// Drop-in replacemnt for go-appimage for sandboxing and use with shappimages
// NOT FINISHED AND STILL LACKING BASIC FEATURES
// THIS SHOULD BE USED FOR TESTING PURPOSES *ONLY* UNTIL IN A STABLE STATE

package main

import "C"
import (
	aisap "github.com/mgord9518/aisap"
)

var openAppImages []*aisap.AppImage

func main() {}

// -------------------- appimage.go ------------------

//export aisap_appimage_new
func aisap_appimage_new(cAi *C.aisap_appimage, src *C.char) int {
	ai, err := aisap.NewAppImage(C.GoString(src))

	if err != nil {
		return errToInt(err)
	}

	// Set all fields
	cAi._index    = C.uint(len(openAppImages))
	cAi.path      = C.CString(ai.Path)
	cAi.name      = C.CString(ai.Name)
	cAi.data_dir  = C.CString(ai.DataDir())
	cAi.root_dir  = C.CString(ai.RootDir())
	cAi.mount_dir = C.CString(ai.MountDir())
	cAi.md5       = C.CString(ai.Md5())
	cAi.run_id    = C.CString(ai.RunId())
	cAi.ai_type   = C.int(ai.Type())

	openAppImages = append(openAppImages, ai)

	return 0
}

//export aisap_new_appimage
func aisap_new_appimage(cAi *C.aisap_appimage, src *C.char) int {
    return aisap_appimage_new(cAi, src)
}

// Just a way to pass the wrap args to the Zig implementation until I can
// re-implement AppImage.WrapArgs as well
//export aisap_appimage_wraparg_next
func aisap_appimage_wraparg_next(cAi *C.aisap_appimage, length *int) *C.char {
	ai := openAppImages[cAi._index]

	var ret *C.char

	if ai.WrapArgsList == nil {
		ai.WrapArgsList, _ = ai.WrapArgs([]string{})
	}

	// Set the return value and iterate the counter
	if ai.CurrentArg < len(ai.WrapArgsList) {
		ret = C.CString(ai.WrapArgsList[ai.CurrentArg])
		*length = len(ai.WrapArgsList[ai.CurrentArg])

		ai.CurrentArg++
	}


	return ret
}

// -------------- wrap.go -----------------

//export aisap_appimage_run
// TODO: Make char get passed correctly. This may just be easiest to just
// make another AppImage run function that accepts **char instead of Go strings
func aisap_appimage_run(cAi *C.aisap_appimage, args **C.char) int {
	return errToInt(openAppImages[cAi._index].Run([]string{}))
}

//export aisap_appimage_sandbox
func aisap_appimage_sandbox(cAi *C.aisap_appimage, args **C.char) int {
	// Set elements of parent Go struct before running
	// They won't be properly applied otherwise
	openAppImages[cAi._index].SetDataDir(C.GoString(cAi.data_dir))
	openAppImages[cAi._index].SetRootDir(C.GoString(cAi.root_dir))

	return errToInt(openAppImages[cAi._index].Sandbox([]string{}))
}

// ------------------- mount.go ---------------

//export aisap_appimage_mount
func aisap_appimage_mount(cAi *C.aisap_appimage) int {
	err := openAppImages[cAi._index].Mount()
	cAi.temp_dir = C.CString(openAppImages[cAi._index].TempDir())

	return errToInt(err)
}

//export aisap_appimage_destroy
func aisap_appimage_destroy(cAi *C.aisap_appimage) {
	openAppImages[cAi._index].Destroy()
	cAi = nil
}

//export aisap_appimage_ismounted
func aisap_appimage_ismounted(cAi *C.aisap_appimage) int {
	if openAppImages[cAi._index].IsMounted() {
		return 1
	}

	return 0
}

func errToInt(err error) int {
	switch err {
	case nil:
		return 0
	case aisap.NilAppImage:
		return 2
	case aisap.NoPath:
		return 3
	case aisap.NotMounted:
		return 4
	case aisap.InvalidDesktopFile:
		return 5
	case aisap.NoIcon:
		return 6
	case aisap.InvalidIconExtension:
		return 7
	case aisap.NoMountPoint:
		return 8
	// Unknown aisap error; return 1
	// TODO: add non-aisap errors, such as os.Open
	default:
		return 1
	}
}
