// Drop-in replacemnt for go-appimage for sandboxing and use with shappimages
// NOT FINISHED AND STILL LACKING BASIC FEATURES
// THIS SHOULD BE USED FOR TESTING PURPOSES *ONLY* UNTIL IN A STABLE STATE

package main

/*
#include <stdlib.h>
struct aisap_AppImage {
	char* name;
	char* path;
	char* data_dir;
	char* root_dir;
	char* temp_dir;
	char* mount_dir;
	char* md5;
	char* run_id;
	unsigned int _index;
	void*        _parent;
	int ai_type;
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

//export aisap_appimage_new
func aisap_appimage_new(cAi *C.aisap_AppImage, src *C.char) int {
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
func aisap_new_appimage(cAi *C.aisap_AppImage, src *C.char) int {
    return aisap_appimage_new(cAi, src)
}

//export aisap_appimage_thumbnail
// TODO: add functionality
//func aisap_appimage_thumbnail(cAi *C.aisap_AppImage) *C.char {
//	return C.CString(openAppImages[cAi._index].Md5())
//}

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
	// Set elements of parent Go struct before running
	// They won't be properly applied otherwise
	openAppImages[cAi._index].SetDataDir(C.GoString(cAi.data_dir))
	openAppImages[cAi._index].SetRootDir(C.GoString(cAi.root_dir))

	return errToInt(openAppImages[cAi._index].Sandbox([]string{}))
}

// ------------------- mount.go ---------------

//export aisap_appimage_mount
func aisap_appimage_mount(cAi *C.aisap_AppImage) int {
	err := openAppImages[cAi._index].Mount()
	cAi.temp_dir = C.CString(openAppImages[cAi._index].TempDir())

	return errToInt(err)
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
