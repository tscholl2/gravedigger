/*
This is a utility that looks for unused functions/variables
for all code in a given directory.
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cmd/guru/serial"
)

var (
	dir     = "."
	useGuru bool
)

type declaration struct {
	pos  token.Position
	node *ast.Ident
}

func main() {
	flag.BoolVar(&useGuru, "guru", false, "use guru for name referencing (SLOW)")
	flag.CommandLine.Usage = func() {
		fmt.Println(`gravedigger: looks for unused code in a directory. This differs
from other packages in that it takes a directory and lists things that are unused
any where in that directory, including exported things in subpackages/subdirectories.
Example: 'gravedigger'
Example: 'gravedigger .'
Example: 'gravedigger test/'
Options: 
`)
		flag.PrintDefaults()
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

	parse(fileSet, &packages, declarations)

	for _, p := range packages {
		for _, f := range p.Files {
			fmt.Println(fileSet.Position(f.Pos()).Filename)
		}
	}

	// Step 1: go through and mark all declarations

	fmt.Printf("\n---------step 1-------------(mark all declarations)\n\n")

	mark(fileSet, packages, declarations)

	for _, dec := range declarations {
		fmt.Printf("* %s = %s\n", dec.node.Name, dec.pos.String())
	}

	// Step 2: go through and unmark all used functions

	fmt.Printf("\n---------step 2-------------(mark all used declarations)\n\n")

	if useGuru {
		unmarkGuru(fileSet, packages, &declarations)
	} else {
		unmarkFast(fileSet, packages, &declarations)
	}

	// Step 3: return a list of unused functions

	fmt.Printf("\n---------step 3-------------(list all unused declarations)\n\n")

	for _, dec := range declarations {
		fmt.Printf("* %s = %s\n", dec.node.Name, dec.pos)
	}

}

func parse(fileSet *token.FileSet, packages *[]*ast.Package, declarations map[string]declaration) {
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
			*packages = append(*packages, p)
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
						// TODO: check signature instead of name
						if node.Name.Name == "init" || node.Name.Name == "MarshalJSON" || node.Name.Name == "Scan" || node.Name.Name == "Value" || (p.Name == "main" && node.Name.Name == "main") {
							return true
						}
						dec := declaration{fileSet.Position(node.Name.Pos()), node.Name}
						filename, _ := filepath.Abs(dec.pos.Filename)
						declarations[fmt.Sprintf("%s:%d:%d", filename, dec.pos.Line, dec.pos.Column)] = dec
					case *ast.GenDecl:
						for _, spec := range node.Specs {
							switch s := spec.(type) {
							case *ast.ValueSpec:
								for _, n := range s.Names {
									dec := declaration{fileSet.Position(n.Pos()), n}
									filename, _ := filepath.Abs(dec.pos.Filename)
									declarations[fmt.Sprintf("%s:%d:%d", filename, dec.pos.Line, dec.pos.Column)] = dec
								}
							case *ast.TypeSpec:
								dec := declaration{fileSet.Position(s.Name.Pos()), s.Name}
								filename, _ := filepath.Abs(dec.pos.Filename)
								declarations[fmt.Sprintf("%s:%d:%d", filename, dec.pos.Line, dec.pos.Column)] = dec
								/*
									ast.Inspect(s.Type, func(n ast.Node) bool {
										node, ok := n.(*ast.StructType)
										if !ok {
											return true
										}
										for _, node2 := range node.Fields.List {
											for _, node3 := range node2.Names {
												dec := declaration{fileSet.Position(node3.Pos()), node3}
												filename, _ := filepath.Abs(dec.pos.Filename)
												declarations[fmt.Sprintf("%s:%d:%d", filename, dec.pos.Line, dec.pos.Column)] = dec
											}
										}
										return true
									})
								*/
							}
						}
					}
					return true
				})
			}
		}
	}
}

func unmarkFast(fileSet *token.FileSet, packages []*ast.Package, declarations *map[string]declaration) {
	simpleDeclarations := make(map[string]declaration)
	for _, dec := range *declarations {
		simpleDeclarations[dec.node.Name] = dec
	}
	for _, p := range packages {
		for _, f := range p.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				node, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				dec, ok := simpleDeclarations[node.Name]
				if !ok {
					return true
				}
				curpos := fileSet.Position(node.Pos())
				if dec.pos.Filename == curpos.Filename && dec.pos.Line == curpos.Line && dec.pos.Column == curpos.Column {
					return true
				}
				fmt.Println(node.Name)
				fmt.Println("used in: ", curpos.String())
				fmt.Println("defd in: ", dec.pos.String())
				delete(simpleDeclarations, node.Name)
				return true
			})
		}
	}
	*declarations = make(map[string]declaration)
	for _, dec := range simpleDeclarations {
		(*declarations)[dec.pos.String()] = dec
	}
}

func unmarkGuru(fileSet *token.FileSet, packages []*ast.Package, declarations *map[string]declaration) {
	names := make(map[string]int)
	for _, dec := range *declarations {
		i, _ := names[dec.node.Name]
		names[dec.node.Name] = i + 1
	}
	updateNames := func(name string) {
		i, ok := names[name]
		if ok && i > 1 {
			names[name] = i - 1
		} else {
			delete(names, name)
		}
	}
	for _, p := range packages {
		for _, f := range p.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				node, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				_, ok = names[node.Name]
				if !ok {
					return true
				}
				position := fileSet.Position(node.Pos())
				currentFile, _ := filepath.Abs(position.Filename)
				out, err := exec.Command("guru", "-json", "definition", fmt.Sprintf("%s:#%d", currentFile, position.Offset)).Output()
				if err != nil && err.Error() != "exit status 1" {
					panic(err)
				}
				if len(out) == 0 {
					return true
				}
				var def serial.Definition
				if err := json.Unmarshal(out, &def); err != nil {
					panic(err)
				}
				if def.ObjPos == "" {
					return true
				}
				currentPosition := fmt.Sprintf("%s:%d:%d", currentFile, position.Line, position.Column)
				defPosition := def.ObjPos
				if currentPosition == defPosition {
					return true
				}
				dec, ok := (*declarations)[defPosition]
				if !ok {
					return true
				}
				fmt.Println(node.Name)
				fmt.Println("used in: ", position.String())
				fmt.Println("defd in: ", dec.pos.String())
				delete(*declarations, defPosition)
				updateNames(dec.node.Name)
				return true
			})
		}
	}
}
