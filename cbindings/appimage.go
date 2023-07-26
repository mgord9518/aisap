// Drop-in replacemnt for go-appimage for sandboxing and use with shappimages
// NOT FINISHED AND STILL LACKING BASIC FEATURES
// THIS SHOULD BE USED FOR TESTING PURPOSES *ONLY* UNTIL IN A STABLE STATE

package main

/*
// Redefine here as including `aisap.h` causes redefinition issues
typedef struct aisap_appimage {
        const char* name;
		size_t      name_len;
        const char* path;
		size_t      path_len;
        unsigned int _go_index;
        void*        _zig_parent;
} aisap_appimage ;

//typedef struct aisap_appimageperms {
//        int    level;
//        char** files;
//        char** devices;
//        char** sockets;
//} aisap_appimageperms;
*/
import "C"
import (
	aisap "github.com/mgord9518/aisap"
)

var openAppImages []*aisap.AppImage

func main() {}

// -------------------- appimage.go ------------------

//export aisap_appimage_init_go
func aisap_appimage_init_go(cAi *C.aisap_appimage, src *C.char) C.int {
	ai, err := aisap.NewAppImage(C.GoString(src))

	// As the intended return value is a positive int, return errors
	// as negated values
	if err != nil {
		return -errToInt(err)
	}

	openAppImages = append(openAppImages, ai)
	return C.int(len(openAppImages) - 1)

	return 0
}

// Just a way to pass the wrap args to the Zig implementation until I can
// re-implement AppImage.WrapArgs as well
//export aisap_appimage_wraparg_next_go
func aisap_appimage_wraparg_next_go(cAi *C.aisap_appimage, length *C.int) *C.char {
	ai := openAppImages[cAi._go_index]

	var ret *C.char

	if ai.WrapArgsList == nil {
		ai.WrapArgsList, _ = ai.WrapArgs([]string{})
	}

	// Set the return value and iterate the counter
	if ai.CurrentArg < len(ai.WrapArgsList) {
		ret = C.CString(ai.WrapArgsList[ai.CurrentArg])
		*length = C.int(len(ai.WrapArgsList[ai.CurrentArg]))

		ai.CurrentArg++
	}

	return ret
}

// -------------- wrap.go -----------------

//export aisap_appimage_run
// TODO: Make char get passed correctly. This may just be easiest to just
// make another AppImage run function that accepts **char instead of Go strings
func aisap_appimage_run(cAi *C.aisap_appimage, args **C.char) C.int {
	return errToInt(openAppImages[cAi._go_index].Run([]string{}))
}

//export aisap_appimage_sandbox
func aisap_appimage_sandbox(cAi *C.aisap_appimage, args **C.char) C.int {
	ai := openAppImages[cAi._go_index]

	return errToInt(ai.Sandbox([]string{}))
}

// ------------------- mount.go ---------------

//export aisap_appimage_destroy_go
func aisap_appimage_destroy_go(cAi *C.aisap_appimage) {
	openAppImages[cAi._go_index].Destroy()
	cAi = nil
}

//export aisap_appimage_ismounted
func aisap_appimage_ismounted(cAi *C.aisap_appimage) C.int {
	if openAppImages[cAi._go_index].IsMounted() {
		return 1
	}

	return 0
}

func errToInt(err error) C.int {
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
