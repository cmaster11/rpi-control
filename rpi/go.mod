module rpi

go 1.15

require (
	github.com/d2r2/go-i2c v0.0.0-20191123181816-73a8a799d6bc
	github.com/davecgh/go-spew v1.1.1
	github.com/spf13/cobra v1.1.2
)

replace github.com/d2r2/go-i2c v0.0.0-20191123181816-73a8a799d6bc => github.com/googolgl/go-i2c v0.0.5
