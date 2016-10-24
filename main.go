/*
This is a utility that looks for unused functions/variables
for all code in a given directory.
*/
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var (
	dir = "."
)

type declaration struct {
	pos  token.Position
	node *ast.Ident
}

func (d declaration) String() string {
	fn, err := filepath.Abs(d.pos.Filename)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s:%d:%d", fn, d.pos.Line, d.pos.Column)
}

func shortenPath(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	s, err := filepath.Rel(cwd, path)
	if err != nil {
		panic(err)
	}
	return s
}

func main() {
	flag.CommandLine.Usage = func() {
		fmt.Println(`gravedigger: looks for unused code in a directory. This differs
from other packages in that it takes a directory and lists things that are unused
any where in that directory, including exported things in subpackages/subdirectories.
Example: 'gravedigger'
Example: 'gravedigger .'
Example: 'gravedigger test/'
`)
	}
	flag.Parse()
	if flag.NArg() > 0 {
		dir = flag.Args()[0]
	}
	if f, err := os.Stat(dir); err != nil || !f.IsDir() {
		panic(dir + " is not a directory")
	}

	// Step 0: parse all packages and subpackages in this directory

	fmt.Printf("\n---------step 0-------------(find all code in directories)\n\n")

	declarations := make(map[string]declaration)
	packages := []*ast.Package{}
	fileSet := token.NewFileSet()

	parse(fileSet, packages, declarations)

	for _, p := range packages {
		for _, f := range p.Files {
			fmt.Println(fileSet.Position(f.Pos()).Filename)
		}
	}

	// Step 1: go through and mark all declarations

	fmt.Printf("\n---------step 1-------------(mark all declarations)\n\n")

	mark(fileSet, packages, declarations)

	for _, dec := range declarations {
		fmt.Printf("* %s = %s\n", dec.node.Name, dec.pos)
	}

	// Step 2: go through and unmark all used functions

	fmt.Printf("\n---------step 2-------------(mark all used declarations)\n\n")

	unmark(fileSet, packages, declarations)

	// Step 3: return a list of unused functions

	fmt.Printf("\n---------step 3-------------(list all unused declarations)\n\n")

	for _, dec := range declarations {
		fmt.Printf("* %s = %s\n", dec.node.Name, dec.pos)
	}

}

func parse(fileSet *token.FileSet, packages []*ast.Package, declarations map[string]declaration) {
	if err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		pkgs, err := parser.ParseDir(fileSet, filePath, func(info os.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		}, 0)
		if err != nil {
			return err
		}
		for _, p := range pkgs {
			packages = append(packages, p)
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func mark(fileSet *token.FileSet, packages []*ast.Package, declarations map[string]declaration) {
	for _, p := range packages {
		for _, f := range p.Files {
			for _, n := range f.Decls {
				ast.Inspect(n, func(n ast.Node) bool {
					switch node := n.(type) {
					case *ast.FuncDecl:
						if node.Name.Name == "init" || node.Name.Name == "_" || (p.Name == "main" && node.Name.Name == "main") {
							return true
						}
						dec := declaration{fileSet.Position(node.Name.Pos()), node.Name}
						declarations[dec.node.Name] = dec
					case *ast.GenDecl:
						for _, spec := range node.Specs {
							switch s := spec.(type) {
							case *ast.ValueSpec:
								for _, n := range s.Names {
									dec := declaration{fileSet.Position(n.Pos()), n}
									declarations[dec.node.Name] = dec
								}
							case *ast.TypeSpec:
								dec := declaration{fileSet.Position(s.Name.Pos()), s.Name}
								declarations[dec.node.Name] = dec
							}
						}
					}
					return true
				})
			}
		}
	}
}

func unmark(fileSet *token.FileSet, packages []*ast.Package, declarations map[string]declaration) {
	for _, p := range packages {
		for _, f := range p.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				node, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				dec, _ := declarations[node.Name]
				curpos := fileSet.Position(node.Pos())
				if dec.pos.Filename == curpos.Filename && dec.pos.Line == curpos.Line && dec.pos.Column == curpos.Column {
					return true
				}
				delete(declarations, node.Name)
				return true
			})
		}
	}
}
