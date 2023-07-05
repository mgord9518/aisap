#include <stddef.h>
#include <stdint.h>
#include <unistd.h>

typedef struct aisap_appimage {
	const char*  name;
	size_t       name_len;

	const char*  path;
	size_t       path_len;

	unsigned int _go_index;   // For Go implementation as structs cannot contain Go pointers
	void*        _zig_parent; // For Zig implemenation, points to Zig AppImage
} aisap_appimage ;

// Not yet sure how I'll get the char** fields set. I've been unable to
// properly find a way to pass either Go or Zig slices to C. I know it's
// possible, but haven't had much luck. Maybe I'll make a function that cycles
// through a single permission every time it's called and just return a char*
// until I can figure it out
typedef struct aisap_appimageperms {
	int    level;
	char** files;
	char** devices;
	char** sockets;
} aisap_appimageperms;

typedef uint8_t aisap_error;

#ifdef __cplusplus
extern "C" {
#endif

// Go-implemented C functions. These are on the way out in favor of the Zig
// implementation. The Go version WILL stay and continue being maintained for
// other Go projects, it just won't have C bindings. This is because Go comes
// with a rather large runtime that would bloat C programs trying to use it,
// along with having some CGo quirks that are annoying to get around that just
// work in Zig. 
// TODO: Make char get passed correctly. This may just be easiest to just
// make another AppImage run function that accepts char** instead of Go strings
extern int   aisap_appimage_init_go(aisap_appimage* ai, const char* src); // Returns index
extern void  aisap_appimage_destroy_go(aisap_appimage* ai);
extern int   aisap_appimage_run(aisap_appimage* ai, char** args);
extern int   aisap_appimage_mount(aisap_appimage* ai);
extern int   aisap_appimage_ismounted(aisap_appimage* ai);

// Zig-implemented C functions
// `aisap_appimage_new` initializes both the Zig and Go AppImage structs, so
// until I can get the rest of the functions ported over you'll still be able
// to call all of them
extern aisap_appimage aisap_appimage_new(const char* src, aisap_error* err);
extern void           aisap_appimage_destroy(aisap_appimage* ai);
extern int8_t         aisap_appimage_type(aisap_appimage* ai, aisap_error* err);
extern size_t         aisap_appimage_offset(aisap_appimage* ai, aisap_error* err);
extern const char*    aisap_appimage_md5(aisap_appimage* ai, char* buf, size_t buf_len, aisap_error* err);

// THESE FUNCTIONS NOT YET IMPLEMENTED
//extern uint8_t aisap_appimage_sandbox(aisap_appimage* ai, int argc, char** args);
//extern char*   aisap_appimage_mountdir(aisap_appimage* ai);
//extern char*   aisap_appimage_tempdir(aisap_appimage* ai);
//extern char*   aisap_appimage_runid(aisap_appimage* ai);
//extern char*   aisap_appimage_wrapargs(aisap_appimage* ai);

// For ABI compat with libAppImage
extern off_t appimage_get_payload_offset(const char* path);
extern int   appimage_get_type(const char* path);
extern char* appimage_get_md5(const char* path);

#ifdef __cplusplus
}
#endif
