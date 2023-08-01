module github.com/mgord9518/aisap/cmd/aisap

go 1.18

replace github.com/mgord9518/aisap => ../../

replace github.com/mgord9518/aisap/permissions => ../../permissions

replace github.com/mgord9518/aisap/profiles => ../../profiles

replace github.com/mgord9518/aisap/spooky => ../../spooky

replace github.com/mgord9518/aisap/helpers => ../../helpers

require (
	github.com/gookit/color v1.5.4
	github.com/mgord9518/aisap v0.0.0-00010101000000-000000000000
	github.com/mgord9518/aisap/helpers v0.0.0-20230730123911-bc6ec574def8
	github.com/mgord9518/aisap/permissions v0.0.0-20230730123911-bc6ec574def8
	github.com/mgord9518/aisap/spooky v0.0.0-20220316134932-8de512d335b0
	github.com/mgord9518/cli v0.0.0-20220722070617-b0ebed6351fe
	github.com/spf13/pflag v1.0.5
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/CalebQ42/fuse v0.1.0 // indirect
	github.com/CalebQ42/squashfs v0.8.1 // indirect
	github.com/adrg/xdg v0.4.0 // indirect
	github.com/klauspost/compress v1.16.4 // indirect
	github.com/mgord9518/aisap/profiles v0.0.0-20230730123911-bc6ec574def8 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/rasky/go-lzo v0.0.0-20200203143853-96a758eda86e // indirect
	github.com/seaweedfs/fuse v1.2.2 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/ulikunitz/xz v0.5.11 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/sys v0.10.0 // indirect
)
