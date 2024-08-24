// Package generator acts as a prisma generator
package generator

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fragdanceone/prisma-client-go/binaries"
	"github.com/fragdanceone/prisma-client-go/binaries/bindata"
	"github.com/fragdanceone/prisma-client-go/binaries/platform"
	"github.com/fragdanceone/prisma-client-go/logger"
)

const DefaultPackageName = "db"

func addDefaults(input *Root) {
	if input.Generator.Config.Package == "" {
		input.Generator.Config.Package = DefaultPackageName
	}
}

// Run invokes the generator, which builds the templates and writes to the specified output file.
func Run(input *Root) error {
	addDefaults(input)

	if input.Generator.Config.DisableGitignore != "true" && input.Generator.Config.DisableGoBinaries != "true" {
		fmt.Println("writing gitignore file")
		// generate a gitignore into the folder
		var gitignore = "# gitignore generated by Prisma Client Go. DO NOT EDIT.\n*_gen.go\n"
		if err := os.WriteFile(path.Join(input.Generator.Output.Value, ".gitignore"), []byte(gitignore), 0644); err != nil {
			return fmt.Errorf("could not write .gitignore: %w", err)
		}
	}

	if err := generateClient(input); err != nil {
		fmt.Printf("generate client: %w", err)
		return fmt.Errorf("generate client: %w", err)
	}

	if err := generateBinaries(input); err != nil {
		return fmt.Errorf("generate binaries: %w", err)
	}

	return nil
}

func generateClient(input *Root) error {
	var buf bytes.Buffer

	ctx := build.Default
	pkg, err := ctx.Import("github.com/fragdanceone/prisma-client-go", ".", build.FindOnly)
	if err != nil {
		return fmt.Errorf("could not get main template asset: %w", err)
	}

	var templates []*template.Template

	templateDir := pkg.Dir + "/generator/templates"
	err = filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".gotpl") {
			tpl, err := template.ParseFiles(path)
			if err != nil {
				return err
			}
			templates = append(templates, tpl.Templates()...)
		}

		return err
	})

	if err != nil {
		return fmt.Errorf("could not walk dir %s: %w", templateDir, err)
	}

	// Run header template first
	header, err := template.ParseFiles(templateDir + "/_header.gotpl")
	if err != nil {
		return fmt.Errorf("could not find header template %s: %w", templateDir, err)
	}

	if err := header.Execute(&buf, input); err != nil {
		return fmt.Errorf("could not write header template: %w", err)
	}
			output := input.Generator.Output.Value
outFile := path.Join(output, "db_gen.go")
	fmt.Println("Process templates")
	// Then process all remaining templates
	for _, tpl := range templates {
		if strings.Contains(tpl.Name(), "_") {
			continue
		}
		fmt.Printf("Processing " + tpl.Name())
		buf.Write([]byte(fmt.Sprintf("// --- template %s ---\n", tpl.Name())))

		if err := tpl.Execute(&buf, input); err != nil {
			return fmt.Errorf("could not write template file %s: %w", tpl.Name(), err)
		}

		if _, err := format.Source(buf.Bytes()); err != nil {
		os.WriteFile(outFile, buf.Bytes(), 0644)
			return fmt.Errorf("could not format source from file %s: %w", tpl.Name(), err)
		}
		fmt.Println(" processed")
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("could not format final source: %w", err)
	}



	if strings.HasSuffix(output, ".go") {
		return fmt.Errorf("generator output should be a directory")
	}

	if err := os.MkdirAll(output, os.ModePerm); err != nil {
		return fmt.Errorf("could not run MkdirAll on path %s: %w", output, err)
	}

	// TODO make this configurable
	
	if err := os.WriteFile(outFile, formatted, 0644); err != nil {
		return fmt.Errorf("could not write template data to file writer %s: %w", outFile, err)
	}

	return nil
}

func generateBinaries(input *Root) error {
	if input.Generator.Config.DisableGoBinaries == "true" {
		return nil
	}

	if input.GetEngineType() == "dataproxy" {
		logger.Debug.Printf("using data proxy; not fetching any engines")
		return nil
	}

	var targets []string

	for _, target := range input.Generator.BinaryTargets {
		targets = append(targets, target.Value)
	}

	targets = add(targets, "native")
	targets = add(targets, "linux")

	// TODO refactor
	for _, name := range targets {
		if name == "native" {
			name = platform.BinaryPlatformName()
		}

		// first, ensure they are actually downloaded
		if err := binaries.FetchEngine(binaries.GlobalCacheDir(), "query-engine", name); err != nil {
			return fmt.Errorf("failed fetching binaries: %w", err)
		}
	}

	if err := generateQueryEngineFiles(targets, input.Generator.Config.Package.String(), input.Generator.Output.Value); err != nil {
		return fmt.Errorf("could not write template data: %w", err)
	}

	return nil
}

func generateQueryEngineFiles(binaryTargets []string, pkg, outputDir string) error {
	for _, name := range binaryTargets {
		if name == "native" {
			name = platform.BinaryPlatformName()
		}

		enginePath := binaries.GetEnginePath(binaries.GlobalCacheDir(), "query-engine", name)

		pt := name
		if strings.Contains(name, "debian") || strings.Contains(name, "rhel") {
			pt = "linux"
		}

		filename := fmt.Sprintf("query-engine-%s_gen.go", name)
		to := path.Join(outputDir, filename)

		// TODO check if already exists, but make sure version matches
		if err := bindata.WriteFile(strings.ReplaceAll(name, "-", "_"), pkg, pt, enginePath, to); err != nil {
			return fmt.Errorf("generate write go file: %w", err)
		}

		logger.Debug.Printf("write go file at %s", filename)
	}

	return nil
}

func add(list []string, item string) []string {
	keys := make(map[string]bool)
	if _, ok := keys[item]; !ok {
		keys[item] = true
		list = append(list, item)
	}
	return list
}
