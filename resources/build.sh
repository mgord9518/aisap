#!/bin/sh
# Quick and dirty shell script to build aisap AppDir from source

[ -z "$ARCH" ] && ARCH=$(uname -m)

if command -v 'mkappimage.AppImage'; then
	aitool() {
		'mkappimage.AppImage' "$@"
	}
elif command -v "mkappimage-$ARCH.AppImage"; then
	aitool() {
		"mkappimage-$ARCH.AppImage" "$@"
	}
elif command -v "mkappimage-649-$ARCH.AppImage"; then
	aitool() {
		"mkappimage-649-$ARCH.AppImage" "$@"
	}
elif command -v 'mkappimage'; then
	aitool() {
		'mkappimage' "$@"
	}
elif command -v "$PWD/mkappimage"; then
	aitool() {
		"$PWD/mkappimage" "$@"
	}
else
	# Hacky one-liner to get the URL to download the latest mkappimage
	mkAppImageUrl=$(curl -q https://api.github.com/repos/probonopd/go-appimage/releases | grep $(uname -m) | grep mkappimage | grep browser_download_url | cut -d'"' -f4 | head -n1)
	echo 'Downloading `mkappimage`'
	wget "$mkAppImageUrl" -O 'mkappimage'
	chmod +x 'mkappimage'
	aitool() {
		"$PWD/mkappimage" "$@"
	}
fi

if [ ! $(command -v 'go') ]; then
	echo 'Failed to locate GoLang compiler! Unable to build'
	exit 1
fi

aisapUrl='github.com/mgord9518/aisap'
aisapRawUrl='raw.githubusercontent.com/mgord9518/aisap/main'

mkdir -p 'AppDir/usr/bin' \
         'AppDir/usr/share/metainfo' \
         'AppDir/usr/share/icons/hicolor/scalable/apps'

# Compile the binary into the AppDir
#CGO_ENABLED=0 GOBIN="$PWD/AppDir/usr/bin" go install -ldflags '-s -w' \
#	"$aisapUrl/aisap-bin@latest"
cd cmd/aisap

echo 'replace github.com/mgord9518/aisap => ../../
replace github.com/mgord9518/aisap/permissions => ../../permissions
replace github.com/mgord9518/aisap/profiles => ../../profiles
replace github.com/mgord9518/aisap/spooky => ../../spooky
replace github.com/mgord9518/aisap/helpers => ../../helpers
' >> go.mod

go mod tidy

CGO_ENABLED=0 go build -ldflags '-s -w' -o '../../AppDir/usr/bin'
[ $? -ne 0 ] && exit $?
cd ../..

# Download icon
wget "$aisapRawUrl/resources/aisap.svg" -O \
	'AppDir/usr/share/icons/hicolor/scalable/apps/io.github.mgord9518.aisap.svg'
[ $? -ne 0 ] && exit $?

# Download desktop entry
wget "$aisapRawUrl/resources/aisap.desktop" -O 'AppDir/io.github.mgord9518.aisap.desktop'
[ $? -ne 0 ] && exit $?

# Download AppStream metainfo
wget "$aisapRawUrl/resources/aisap.appdata.xml" -O \
	'AppDir/usr/share/metainfo/io.github.mgord9518.aisap.appdata.xml'
[ $? -ne 0 ] && exit $?

# Download squashfuse binary
wget "https://github.com/mgord9518/portable_squashfuse/releases/download/nightly/squashfuse_lz4_xz_zstd-static.$ARCH" -O 'AppDir/usr/bin/squashfuse'
chmod +x 'AppDir/usr/bin/squashfuse'
[ $? -ne 0 ] && exit $?

# Download bwrap binary
wget "https://github.com/mgord9518/portable_bwrap/releases/download/nightly/bwrap-static.$ARCH" -O 'AppDir/usr/bin/bwrap'
chmod +x 'AppDir/usr/bin/bwrap'
[ $? -ne 0 ] && exit $?

# Download excludelist
wget 'https://raw.githubusercontent.com/AppImage/pkg2appimage/master/excludelist' -O \
	'excludelist'

# Link up files
ln -s './usr/share/icons/hicolor/scalable/apps/io.github.mgord9518.aisap.svg' 'AppDir/io.github.mgord9518.aisap.svg'
ln -s './usr/bin/aisap' 'AppDir/AppRun'

# Build the AppImage
export ARCH="$ARCH"
export VERSION=$('AppDir/usr/bin/aisap' --version)

sed -i 's/X-AppImage-Architecture.*/X-AppImage-Architecture=x86_64/' 'AppDir/io.github.mgord9518.aisap.desktop'

aitool -u "gh-releases-zsync|mgord9518|aisap|continuous|aisap-*$ARCH.AppImage.zsync" AppDir
[ $? -ne 0 ] && exit $?

# Build for ARM
#cd aisap-bin
#CGO_ENABLED=0 GOARCH=arm GOARM=5 go build -ldflags '-s -w' -o '../AppDir/usr/bin'
#cd ..
#
## Download squashfuse binary
#wget "https://github.com/mgord9518/portable_squashfuse/releases/download/continuous/squashfuse_lz4_xz_zstd.$ARCH" -O 'AppDir/usr/bin/squashfuse'
#chmod +x 'AppDir/usr/bin/squashfuse'
#
#export ARCH="armhf"
#aitool -u "gh-releases-zsync|mgord9518|aisap|continuous|aisap-*$ARCH.AppImage.zsync" AppDir

# Experimental multi-arch shImg build
mkdir -p 'AppDir/usr.aarch64/bin'
cd cmd/aisap
go mod tidy
CGO_ENABLED=0 GOARCH=arm64 go build -ldflags '-s -w' -o '../../AppDir/usr.aarch64/bin'
cd ../..
ln -s './usr.aarch64/bin/aisap' 'AppDir/AppRun.aarch64'

# Download squashfuse binary
wget "https://github.com/mgord9518/portable_squashfuse/releases/download/nightly/squashfuse_lz4_xz_zstd-static.aarch64" -O 'AppDir/usr.aarch64/bin/squashfuse'
chmod +x 'AppDir/usr.aarch64/bin/squashfuse'
[ $? -ne 0 ] && exit $?

# Download bwrap binary
wget "https://github.com/mgord9518/portable_bwrap/releases/download/nightly/bwrap-static.$ARCH" -O 'AppDir/usr.aarch64/bin/bwrap'
chmod +x 'AppDir/usr.aarch64/bin/bwrap'
[ $? -ne 0 ] && exit $?


sed -i 's/X-AppImage-Architecture.*/X-AppImage-Architecture=x86_64;aarch64/' 'AppDir/io.github.mgord9518.aisap.desktop'

# Build SquashFS image
mksquashfs AppDir sfs -root-owned -no-exports -noI -b 1M -comp lz4 -Xhc -nopad
[ $? -ne 0 ] && exit $?

# Download shImg runtime
wget "https://github.com/mgord9518/shappimage/releases/download/continuous/runtime-lz4-static-x86_64-aarch64"
[ $? -ne 0 ] && exit $?

cat runtime-lz4-static-x86_64-aarch64 sfs > "aisap-$VERSION-x86_64_aarch64.shImg"
chmod +x "aisap-$VERSION-x86_64_aarch64.shImg"

# Append desktop integration info
wget 'https://raw.githubusercontent.com/mgord9518/shappimage/main/add_integration.sh'
[ $? -ne 0 ] && exit $?
sh add_integration.sh ./"aisap-$VERSION-x86_64_aarch64.shImg" "gh-releases-zsync|mgord9518|aisap|continuous|aisap-*-x86_64_aarch64.shImg.zsync"

exit 0
