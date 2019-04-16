package main

import (
	_ "github.com/bblfsh/python-driver/driver/impl"
	"github.com/bblfsh/python-driver/driver/normalizer"

	"github.com/bblfsh/sdk/v3/driver/server"
)

func main() {
	server.Run(normalizer.Transforms)
}
