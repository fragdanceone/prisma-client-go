//go:build tools
// +build tools

// See https://github.com/golang/go/issues/26366
package generator

import (
	_ "github.com/fragdance1/prisma-client-go/generator/templates"
	_ "github.com/fragdance1/prisma-client-go/generator/templates/actions"
)
