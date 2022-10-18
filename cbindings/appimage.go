// Drop-in replacemnt for go-appimage for sandboxing and use with shappimages
// NOT FINISHED AND STILL LACKING BASIC FEATURES
// THIS SHOULD BE USED FOR TESTING PURPOSES *ONLY* UNTIL IN A STABLE STATE

package main

/*
struct aisap_AppImage {
	char* name;
	char* path;
	unsigned int _index;
};
struct aisap_AppImagePerms {
	int    level;
	char** files;
	char** devices;
	char** sockets;
};
typedef struct aisap_AppImage aisap_AppImage;
typedef struct aisap_AppImagePerms aisap_AppImagePerms;
*/
import "C"
import (
	aisap "github.com/mgord9518/aisap"
)

var openAppImages []*aisap.AppImage

func main() {}

// -------------------- appimage.go ------------------
// TODO: Finish appimage.go funcs

//export aisap_new_appimage
func aisap_new_appimage(cAi *C.aisap_AppImage, src *C.char) int {
	ai, err := aisap.NewAppImage(C.GoString(src))

	cAi._index = C.uint(len(openAppImages))
	cAi.name = C.CString(ai.Name)

	openAppImages = append(openAppImages, ai)

	return errToInt(err)
}

//export aisap_appimage_thumbnail
// TODO: add functionality
//func aisap_appimage_thumbnail(cAi *C.aisap_AppImage) *C.char {
//	return C.CString(openAppImages[cAi._index].Md5())
//}

//export aisap_appimage_md5
func aisap_appimage_md5(cAi *C.aisap_AppImage) *C.char {
	return C.CString(openAppImages[cAi._index].Md5())
}

//export aisap_appimage_tempdir
func aisap_appimage_tempdir(cAi *C.aisap_AppImage) *C.char {
	return C.CString(openAppImages[cAi._index].TempDir())
}

//export aisap_appimage_mountdir
func aisap_appimage_mountdir(cAi *C.aisap_AppImage) *C.char {
	return C.CString(openAppImages[cAi._index].MountDir())
}

//export aisap_appimage_runid
func aisap_appimage_runid(cAi *C.aisap_AppImage) *C.char {
	return C.CString(openAppImages[cAi._index].RunId())
}

//export aisap_appimage_set_rootdir
func aisap_appimage_set_rootdir(cAi *C.aisap_AppImage, d *C.char) {
	openAppImages[cAi._index].SetRootDir(C.GoString(d))
}

//export aisap_appimage_set_datadir
func aisap_appimage_set_datadir(cAi *C.aisap_AppImage, d *C.char) {
	openAppImages[cAi._index].SetDataDir(C.GoString(d))
}

//export aisap_appimage_set_tempdir
func aisap_appimage_set_tempdir(cAi *C.aisap_AppImage, d *C.char) {
	openAppImages[cAi._index].SetTempDir(C.GoString(d))
}

//export aisap_appimage_type
func aisap_appimage_type(cAi *C.aisap_AppImage) int {
	return openAppImages[cAi._index].Type()
}

//export aisap_appimage_archetectures
// TODO: add functionality
//func aisap_appimage_archetectures(cAi *C.aisap_AppImage) **C.char {
//	return openAppImages[cAi._index].Archetectures()
//}

// -------------- wrap.go -----------------

//export aisap_appimage_run
// TODO: Make char get passed correctly. This may just be easiest to just
// make another AppImage run function that accepts **char instead of Go strings
func aisap_appimage_run(cAi *C.aisap_AppImage, args **C.char) int {
	return errToInt(openAppImages[cAi._index].Run([]string{}))
}

//export aisap_appimage_sandbox
func aisap_appimage_sandbox(cAi *C.aisap_AppImage, args **C.char) int {
	return errToInt(openAppImages[cAi._index].Sandbox([]string{}))
}

//export aisap_appimage_wrap_args
// TODO: add functionality
//func aisap_appimage_wrap_args(cAi *C.aisap_AppImage, args **C.char) **C.char {
	//return errToInt(openAppImages[cAi._index].Sandbox([]string{}))
//}

// ------------------- mount.go ---------------

//export aisap_appimage_mount
func aisap_appimage_mount(cAi *C.aisap_AppImage) int {
	return errToInt(openAppImages[cAi._index].Mount())
}

//export aisap_appimage_destroy
func aisap_appimage_destroy(cAi *C.aisap_AppImage) {
	openAppImages[cAi._index].Destroy()
	cAi = nil
}

//export aisap_appimage_ismounted
func aisap_appimage_ismounted(cAi *C.aisap_AppImage) int {
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
