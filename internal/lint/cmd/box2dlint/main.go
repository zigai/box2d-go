package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/zigai/box2d-go/internal/lint/fmaround"
)

func main() { singlechecker.Main(fmaround.Analyzer) }
