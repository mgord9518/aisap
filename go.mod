module github.com/mgord9518/aisap

go 1.16

replace github.com/mgord9518/aisap/helpers => ./helpers

replace github.com/mgord9518/aisap/profiles => ./profiles

require (
	github.com/adrg/xdg v0.3.4
	github.com/mgord9518/aisap/helpers v0.0.0-00010101000000-000000000000
	github.com/mgord9518/aisap/profiles v0.0.0-00010101000000-000000000000
	github.com/smartystreets/goconvey v1.7.2 // indirect
	gopkg.in/ini.v1 v1.63.0
)
