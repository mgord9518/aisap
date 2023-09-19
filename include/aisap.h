#include <stddef.h>
#include <stdint.h>
#include <unistd.h>
#include <stdbool.h>

typedef struct aisap_appimage {
	const char*  name;
	size_t       name_len;

	const char*  path;
	size_t       path_len;

	// For Zig implemenation, points to Zig AppImage
	void*        _zig_parent; 
} aisap_appimage ;

typedef enum aisap_socket {
	AISAP_SOCKET_ALSA,
	AISAP_SOCKET_AUDIO,
	AISAP_SOCKET_CGROUP,
	AISAP_SOCKET_DBUS,
	AISAP_SOCKET_IPC,
	AISAP_SOCKET_NETWORK,
	AISAP_SOCKET_PID,
	AISAP_SOCKET_PIPEWIRE,
	AISAP_SOCKET_PULSEAUDIO,
	AISAP_SOCKET_SESSION,
	AISAP_SOCKET_USER,
	AISAP_SOCKET_UTS,
	AISAP_SOCKET_WAYLAND,
	AISAP_SOCKET_X11,
} aisap_socket;

typedef struct aisap_file {
	// The file's real path
	const char*	  src_path;
	size_t        src_path_len;

	// The path it'll be exposed to in the sandbox
	const char*	  dest_path;
	size_t        dest_path_len;

	bool writable;
} aisap_file;

// Not yet implemented, just messing around with potential API at the moment
typedef struct aisap_permissions {
	// level ranges from 0 to 3
	uint8_t		  level;

	aisap_file*	  files;
	size_t	   	  files_len;

	char**		  devices;
	size_t	   	  devices_len;

	aisap_socket* sockets;
	size_t	   	  sockets_len;
} aisap_permissions;

typedef enum aisap_bundle_type {
	AISAP_BUNDLE_SHIMG = -2,
	AISAP_BUNDLE_TYPE1 = 1,
	AISAP_BUNDLE_TYPE2 = 2,
} aisap_bundle_type;

typedef uint8_t aisap_error;

#ifdef __cplusplus
extern "C" {
#endif

// Zig-implemented C functions
// `aisap_appimage_new` initializes both the Zig and Go AppImage structs, so
// until I can get the rest of the functions ported over you'll still be able
// to call all of them
extern aisap_appimage aisap_appimage_new(const char* path, aisap_error* err);

// Like `aisap_appimage_new`, but takes a path length instead of a
// null-terminated path. For this function, path does NOT have to be null-
// terminated
extern aisap_appimage aisap_appimage_newn(const char* path, size_t path_len, aisap_error* err);

extern void              aisap_appimage_destroy(aisap_appimage* ai);
extern aisap_bundle_type aisap_appimage_type(aisap_appimage* ai, aisap_error* err);
extern size_t            aisap_appimage_offset(aisap_appimage* ai, aisap_error* err);
extern const char*       aisap_appimage_md5(aisap_appimage* ai, char* buf, size_t buf_len, aisap_error* err);
extern void              aisap_appimage_mount(aisap_appimage* ai, char* path, aisap_error* err);
extern const char*       aisap_appimage_mount_dir(aisap_appimage* ai);
extern void              aisap_appimage_sandbox(aisap_appimage* ai, int argc, char** args, aisap_error* err);
extern char**            aisap_appimage_wrapargs(aisap_appimage* ai, aisap_error* err);

// THESE FUNCTIONS NOT YET IMPLEMENTED
//extern char*   aisap_appimage_mountdir(aisap_appimage* ai);
//extern char*   aisap_appimage_tempdir(aisap_appimage* ai);
//extern char*   aisap_appimage_runid(aisap_appimage* ai);

// For ABI compat with libAppImage
extern off_t appimage_get_payload_offset(const char* path);
extern int   appimage_get_type(const char* path);
extern char* appimage_get_md5(const char* path);

#ifdef __cplusplus
}
#endif
