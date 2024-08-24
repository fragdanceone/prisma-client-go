package bindata

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/fragdanceone/prisma-client-go/binaries"
)

// TODO go fmt files after creation

func WriteFile(name, pkg, platform, from, to string) error {
	f, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("generate open go file: %w", err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer f.Close()

	if err := writeHeader(f, pkg, name, platform); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	if err := writeAsset(f, from); err != nil {
		return fmt.Errorf("write asset: %w", err)
	}

	return nil
}

func writeHeader(w io.Writer, pkg, name, platform string) error {
	var constraints string
	if platform == "linux" {
		if name == "linux" {
			// TODO dynamically construct these with allTargets in run.go
			// TODO only include these for engines, not for the CLI
			constraints = `// +build !debian_openssl_1_0_x
// +build !debian_openssl_1_1_x
// +build !rhel_openssl_1_0_x
// +build !rhel_openssl_1_1_x`
			constraints += "\n"
		} else {
			constraints = "// +build linux\n"
		}
	}

	_, err := fmt.Fprintf(w, `// Code generated by Prisma Client Go. DO NOT EDIT.
//nolint
// +build !codeanalysis
// +build %s
// +build !prisma_ignore
%s
package %s

import (
	"github.com/fragdanceone/prisma-client-go/binaries/unpack"
)

func init() {
	unpack.Unpack(data, "%s", "%s")
}
`, name, constraints, pkg, name, binaries.EngineVersion)
	return err
}

func writeAsset(w io.Writer, file string) error {
	fd, err := os.Open(file)
	if err != nil {
		return err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer fd.Close()

	h := sha256.New()
	tr := io.TeeReader(fd, h)

	if err := uncompressedMemcopy(w, tr); err != nil {
		return err
	}

	return nil
}

func uncompressedMemcopy(w io.Writer, r io.Reader) error {
	if _, err := fmt.Fprintf(w, `var data = []byte(`); err != nil {
		return err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%+q", b); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, `)`); err != nil {
		return err
	}
	return nil
}
