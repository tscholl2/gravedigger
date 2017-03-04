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
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	dir = "."
)

func main() {
	flag.CommandLine.Usage = func() {
		fmt.Println(`gravedigger [directory]: looks for unused code in a directory. This differs
from other packages in that it takes a directory and lists things that are unused
any where in that directory, including exported things in subpackages/subdirectories.
Example: 'gravedigger'
Example: 'gravedigger .'
Example: 'gravedigger test/'`)
	}
	flag.Parse()
	if flag.NArg() > 0 {
		dir = flag.Args()[0]
	}
	if f, err := os.Stat(dir); err != nil || !f.IsDir() {
		log.Fatal(dir + ": not a directory")
	}
	fileSet := token.NewFileSet()
	// Step 0: parse all packages and subpackages in this directory
	packages := parse(fileSet)
	// Step 1: go through and mark all declarations
	declarations := mark(packages)
	// Step 2: go through and remove all used functions
	declarations = unmark(packages, declarations)
	// Step 3: print output
	print(fileSet, declarations)
}

func parse(fileSet *token.FileSet) (packages []*ast.Package) {
	filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		pkgs, err := parser.ParseDir(fileSet, filePath, func(info os.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		}, 0)
		if err != nil {
			panic(err)
		}
		for _, p := range pkgs {
			packages = append(packages, p)
		}
		return nil
	})
	return packages
}

func mark(packages []*ast.Package) (declarations []*ast.Ident) {
	for _, p := range packages {
		for _, f := range p.Files {
			for _, n := range f.Decls {
				ast.Inspect(n, func(n ast.Node) bool {
					switch node := n.(type) {
					case *ast.FuncDecl:
						if node.Name.Name == "init" || node.Name.Name == "MarshalJSON" || node.Name.Name == "Scan" || node.Name.Name == "Value" || (p.Name == "main" && node.Name.Name == "main") {
							return true
						}
						declarations = append(declarations, node.Name)
					case *ast.GenDecl:
						for _, spec := range node.Specs {
							switch s := spec.(type) {
							case *ast.ValueSpec:
								for _, n := range s.Names {
									declarations = append(declarations, n)
								}
							case *ast.TypeSpec:
								declarations = append(declarations, s.Name)
								ast.Inspect(s.Type, func(n ast.Node) bool {
									node, ok := n.(*ast.StructType)
									if !ok {
										return true
									}
									for _, node2 := range node.Fields.List {
										for _, node3 := range node2.Names {
											declarations = append(declarations, node3)
										}
									}
									return true
								})
							}
						}
					}
					return true
				})
			}
		}
	}
	return declarations
}

func unmark(packages []*ast.Package, declarations []*ast.Ident) (unused []*ast.Ident) {
	unusedDeclarations := make(map[string]*ast.Ident)
	for _, n := range declarations {
		unusedDeclarations[n.Name] = n
	}
	for _, p := range packages {
		for _, f := range p.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				node, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				oldNode, ok := unusedDeclarations[node.Name]
				if !ok {
					return true
				}
				if oldNode.Pos() != node.Pos() {
					delete(unusedDeclarations, node.Name)
				}
				return true
			})
		}
	}
	for _, d := range unusedDeclarations {
		unused = append(unused, d)
	}
	return
}

func print(fileSet *token.FileSet, declarations []*ast.Ident) {
	declarationsByFile := make(map[string][]*ast.Ident)
	for _, n := range declarations {
		if _, ok := declarationsByFile[fileSet.Position(n.Pos()).Filename]; !ok {
			declarationsByFile[fileSet.Position(n.Pos()).Filename] = nil
		}
		declarationsByFile[fileSet.Position(n.Pos()).Filename] = append(declarationsByFile[fileSet.Position(n.Pos()).Filename], n)
	}
	for f := range declarationsByFile {
		fmt.Printf("%s\n", f)
		sort.Slice(declarationsByFile[f], func(i, j int) bool {
			return fileSet.Position(declarationsByFile[f][i].Pos()).Line < fileSet.Position(declarationsByFile[f][j].Pos()).Line
		})
		for _, n := range declarationsByFile[f] {
			fmt.Printf("\t- %s:%d:%d\n", n.Name, fileSet.Position(n.Pos()).Line, fileSet.Position(n.Pos()).Column)
		}
	}
}
