package main

import (
	"github.com/sampila/uniai-client/cmd"
	"github.com/unidoc/unipdf/v4/common/license"
	"os"
)

func init() {
	err := license.SetMeteredKey(os.Getenv("UNIDOC_LICENSE_API_KEY_DEV"))
	if err != nil {
		panic(err)
	}
}

func main() {
	cmd.Execute()
}
