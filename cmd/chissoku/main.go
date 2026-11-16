// Package chissoku implements main chissoku program
package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/northeye/chissoku/internal/chissoku"
)

func main() {
	var c chissoku.Chissoku
	ctx := kong.Parse(&c,
		kong.Name(c.ProgramName()),
		kong.Vars{"version": "v" + c.Version(), "outputters": strings.Join(c.RegisterOutputters(), ",")},
		kong.Description(c.Description()),
		kong.Bind(&c.Options))
	if err := ctx.Run(); err != nil {
		slog.Error("chissoku.Run()", "error", err)
		os.Exit(1)
	}
}
