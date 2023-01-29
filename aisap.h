#include <stddef.h>
#include <stdint.h>
#include <unistd.h>

typedef struct aisap_AppImage {
	char* name;
	char* path;
	char* data_dir;
	char* root_dir;
	char* temp_dir;
	char* mount_dir;
	char* md5;
	char* run_id;
	unsigned int _index;  // For Go implementation as structs cannot contain Go pointers
	void*        _parent; // For Zig implemenation, points to Zig AppImage
	int ai_type;
} aisap_AppImage;

// Not yet sure how I'll get the char** fields set. I've been unable to
// properly find a way to pass either Go or Zig slices to C. I know it's
// possible, but haven't had much luck. Maybe I'll make a function that cycles
// through a single permission every time it's called and just return a char*
// until I can figure it out
typedef struct aisap_AppImagePerms {
	int    level;
	char** files;
	char** devices;
	char** sockets;
} aisap_AppImagePerms;

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
extern int   aisap_new_appimage(aisap_AppImage* cAi, char* src); // Probably going to deprecate this as it doesn't follow the naming convention of the rest of the API
extern int   aisap_appimage_new(aisap_AppImage* cAi);            // This will likely replace `aisap_new_appimage`
extern void  aisap_appimage_destroy(aisap_AppImage* cAi);
extern int   aisap_appimage_run(aisap_AppImage* cAi, char** args);
extern int   aisap_appimage_mount(aisap_AppImage* cAi);
extern int   aisap_appimage_ismounted(aisap_AppImage* cAi);

// Zig-implemented C functions
extern int          aisap_appimage_type(aisap_AppImage* cAi);
extern char*        aisap_appimage_md5(aisap_AppImage* cAi);
extern char*        aisap_appimage_mountdir(aisap_AppImage* cAi);
extern char*        aisap_appimage_tempdir(aisap_AppImage* cAi);
extern char*        aisap_appimage_runid(aisap_AppImage* cAi);
extern int          aisap_appimage_sandbox(aisap_AppImage* cAi, int argc, char** args);
extern unsigned int aisap_appimage_offset(aisap_AppImage* ai, unsigned int* off);
extern char*        aisap_appimage_wrapargs(aisap_AppImage* cAi);

// For ABI/API compat with libAppImage
extern off_t appimage_get_payload_offset(char* path);
extern int   appimage_get_type(char* path);
extern char* appimage_get_md5(char* path);

#ifdef __cplusplus
}
#endif
