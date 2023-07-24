 zig cc -static \
	-o test test.c \
	../libaisap-x86_64.a \
#	-I ../ \
#	-I squashfuse-zig/zstd/lib \
#	-DENABLE_ZLIB \
#	-DUSE_LIBDEFLATE \
#	-DENABLE_ZSTD
#	squashfuse-zig/zstd/lib/decompress/zstd_decompress.c \
#	squashfuse-zig/zstd/lib/decompress/zstd_decompress_block.c \
#	squashfuse-zig/zstd/lib/decompress/zstd_ddict.c \
#	squashfuse-zig/zstd/lib/decompress/huf_decompress.c \
#	squashfuse-zig/zstd/lib/common/zstd_common.c \
#	squashfuse-zig/zstd/lib/common/error_private.c \
#	squashfuse-zig/zstd/lib/common/entropy_common.c \
#	squashfuse-zig/zstd/lib/common/fse_decompress.c \
#	squashfuse-zig/zstd/lib/common/xxhash.c \
#	squashfuse-zig/zstd/lib/decompress/huf_decompress_amd64.S \
#	\
#	squashfuse-zig/libdeflate/lib/adler32.c \
#	squashfuse-zig/libdeflate/lib/crc32.c \
#	squashfuse-zig/libdeflate/lib/deflate_decompress.c \
#	squashfuse-zig/libdeflate/lib/utils.c \
#	squashfuse-zig/libdeflate/lib/zlib_decompress.c \
#	squashfuse-zig/libdeflate/lib/x86/cpu_features.c \
#	\
#	squashfuse-zig/lz4/lib/lz4.c \
#	\
#	squashfuse-zig/squashfuse/fs.c \
#	squashfuse-zig/squashfuse/table.c \
#	squashfuse-zig/squashfuse/xattr.c \
#	squashfuse-zig/squashfuse/cache.c \
#	squashfuse-zig/squashfuse/dir.c \
#	squashfuse-zig/squashfuse/file.c \
#	squashfuse-zig/squashfuse/nonstd-makedev.c \
#	squashfuse-zig/squashfuse/nonstd-pread.c \
#	squashfuse-zig/squashfuse/nonstd-stat.c \
#	squashfuse-zig/squashfuse/stat.c \
#	squashfuse-zig/squashfuse/stack.c \
#	squashfuse-zig/squashfuse/swap.c \
