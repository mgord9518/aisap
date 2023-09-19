// Simple C program to test aisap's C bindings
// Build it by running `build_test.sh`.
// This will eventually be in `examples` and built with `build.zig`

#include "aisap.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char** argv) {
	if (argc < 2) {
		printf("usage: %s [appimage]\n", argv[0]);
		return 1;
	}

	aisap_error err;

	aisap_appimage ai = aisap_appimage_new(argv[1], &err);

	if (err) {
		printf("%d\n", err);
		return err;
	}

	// Buf must be at least 33 (32 for hexadecimal MD5, 1 for null byte)
	char buf[33];

	printf("name: %s\n", ai.name);
	printf("path: %s\n", ai.path);

	printf("\n");

	// Test aisap API:
	printf("aisap API:\n");
	printf("  offset: %lu\n", aisap_appimage_offset(&ai, &err));

	const char* md5_string = aisap_appimage_md5(&ai, buf, sizeof(buf), &err);

	if (err) {
		printf("%d\n", err);
		return err;
	}

	printf("  md5:    %s\n",  md5_string);

	printf("  MOUNTING!\n");
	aisap_appimage_mount(&ai, NULL, &err);

	printf("  mount_dir: %s\n", aisap_appimage_mount_dir(&ai));

	char** wrap_args = aisap_appimage_wrapargs(&ai, &err);
	if (err) {
		printf("%d\n", err);
		return err;
	}

	printf("  wrapargs:\n");
    char** i = wrap_args;
    for (char* str = *i; str; str = *++i) {
        printf("%s ", str);
    }
	printf("\n");

    aisap_appimage_sandbox(&ai, argc - 2, argv + 2, &err);
	if (err) {
		printf("%d\n", err);
		return err;
	}

	// Test libappimage API:
	printf("libappimage API:\n");
	printf("  offset: %ld\n", appimage_get_payload_offset(ai.path));
	printf("  md5:    %s\n",  appimage_get_md5(ai.path));

	printf("cleaning up\n");

	aisap_appimage_destroy(&ai);
}
