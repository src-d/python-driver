package main

// FIXME: change "juanjux" to bblfsh
import (
	"github.com/bblfsh/sdk/protocol/cmd"
	"github.com/juanjux/python-driver/driver/normalizer"
)

var version string
var build string

func main() {
	cmd.DriverMain(version, build,
		normalizer.NativeToNoder,
		normalizer.AnnotationRules)
}
