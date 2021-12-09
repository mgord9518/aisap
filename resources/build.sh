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
elif command -v 'appimagetool'; then
	aitool() {
		'appimagetool' "$@"
	}
else
	echo 'Failed to locate appimagetool in $PATH! Unable to build'
	exit 1
fi

if [ ! $(command -v 'go') ]; then
	echo 'Failed to locate GoLang compiler! Unable to build'
	exit 1
fi

aisapUrl='github.com/mgord9518/aisap'
aisapRawUrl='raw.githubusercontent.com/mgord9518/aisap/main'

rm -r 'AppDir' "aisap-$ARCH.AppImage" "aisap-$ARCH.AppImage.zsync"

mkdir -p 'AppDir/usr/bin' \
         'AppDir/usr/share/metainfo' \
         'AppDir/usr/share/icons/hicolor/scalable/apps'

# Download and compile the binary into the AppDir
GOBIN="$PWD/AppDir/usr/bin" go install -ldflags '-s -w' \
	"$aisapUrl/aisap-bin@latest"

if [ $? -ne 0 ]; then
	echo "Failed to build!"
	exit 1
fi

# Download icon
wget "$aisapRawUrl/resources/aisap.svg" -O \
	'AppDir/usr/share/icons/hicolor/scalable/apps/io.github.mgord9518.aisap.svg'

# Download desktop entry
wget "$aisapRawUrl/resources/aisap.desktop" -O 'AppDir/io.github.mgord9518.aisap.desktop'

# Download AppStream metainfo
wget "$aisapRawUrl/resources/aisap.appdata.xml" -O \
	'AppDir/usr/share/metainfo/io.github.mgord9518.aisap.appdata.xml'

# Download squashfuse binary
wget "$aisapRawUrl/resources/squashfuse.$ARCH" -O 'AppDir/usr/bin/squashfuse'
chmod +x 'AppDir/usr/bin/squashfuse'

# Link up files
ln -s './usr/share/icons/hicolor/scalable/apps/io.github.mgord9518.aisap.svg' 'AppDir/io.github.mgord9518.aisap.svg'
ln -s './usr/bin/aisap-bin' 'AppDir/AppRun'

# Build the AppImage
ARCH="$ARCH" VERSION=$('AppDir/usr/bin/aisap-bin' --version) \
	aitool -u "gh-releases-zsync|mgord9518|aisap|continuous|aisap-$ARCH.AppImage.zsync" AppDir
mv 'aisap-'*'.AppImage' "aisap-$ARCH.AppImage"
mv 'aisap-'*'.AppImage.zsync' "aisap-$ARCH.AppImage.zsync"
