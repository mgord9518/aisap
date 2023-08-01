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

	printf("%s\n", ai.name);
	printf("%llu\n", ai.name_len);
	printf("%s\n", ai.path);
	printf("%llu\n", ai.path_len);
	printf("%llu\n", ai._go_index);
	printf("%llu\n", ai._zig_parent);

	aisap_appimage_mount(&ai, "/tmp/ligma", &err);
	sleep(1);

	printf("%llu\n", ai._zig_parent);
	printf("%llu\n", ai._go_index);

	printf("  wrapargs: %s", aisap_appimage_wrapargs(&ai, &err)[0]);
	if (err) {
		printf("%d\n", err);
		return err;
	}

	printf("name: %s\n", ai.name);
	printf("path: %s\n", ai.path);

	printf("\n");

	// Test aisap API:
	printf("aisap API:\n");
	printf("  offset: %lu\n", aisap_appimage_offset(&ai, &err));

	const char* md5_string2 = aisap_appimage_md5(&ai, buf, sizeof(buf), &err);

	if (err) {
		printf("%d\n", err);
		return err;
	}

	printf("  md5:    %s\n",  md5_string2);

	printf("\n");

	// Test libappimage API:
	printf("libappimage API:\n");
	printf("  offset: %ld\n", appimage_get_payload_offset(ai.path));
	printf("  md5:    %s\n",  appimage_get_md5(ai.path));

	aisap_appimage_destroy(&ai);
}
