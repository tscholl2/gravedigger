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
	"strconv"
	"strings"

	"golang.org/x/tools/cmd/guru/serial"
)

var (
	dir = "."
)

type declaration struct {
	pos  token.Position
	name string
}

func (d declaration) String() string {
	fn, err := filepath.Abs(d.pos.Filename)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s:%d:%d", fn, d.pos.Line, d.pos.Column)
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
		flag.CommandLine.PrintDefaults()
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
	if err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		pkgs, err := parser.ParseDir(fileSet, filePath, nil, 0)
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

	for _, p := range packages {
		for _, f := range p.Files {
			fmt.Println(fileSet.Position(f.Pos()).Filename)
		}
	}

	// Step 1: go through and mark all declarations

	fmt.Printf("\n---------step 1-------------(mark all declarations)\n\n")

	for _, p := range packages {
		for _, f := range p.Files {
			for _, n := range f.Decls {
				ast.Inspect(n, func(n ast.Node) bool {
					switch node := n.(type) {
					case *ast.FuncDecl:
						if node.Name.Name == "init" || node.Name.Name == "_" || (p.Name == "main" && node.Name.Name == "main") {
							return true
						}
						dec := declaration{fileSet.Position(node.Name.Pos()), node.Name.Name}
						declarations[dec.String()] = dec
					case *ast.GenDecl:
						// var, const, types
						for _, spec := range node.Specs {
							switch s := spec.(type) {
							case *ast.ValueSpec:
								// constants and variables
								for _, n := range s.Names {
									dec := declaration{fileSet.Position(n.Pos()), n.Name}
									declarations[dec.String()] = dec
								}
							case *ast.TypeSpec:
								// type definitions
								dec := declaration{fileSet.Position(s.Name.Pos()), s.Name.Name}
								declarations[dec.String()] = dec
								// add struct fields?
								ast.Inspect(s.Type, func(n ast.Node) bool {
									node, ok := n.(*ast.StructType)
									if !ok {
										return true
									}
									for _, node2 := range node.Fields.List {
										for _, node3 := range node2.Names {
											dec := declaration{fileSet.Position(node3.Pos()), node3.Name}
											declarations[dec.String()] = dec
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

	for pos, name := range declarations {
		fmt.Printf("* %s = %s\n", name.name, pos)
	}

	// Step 2: go through and unmark all used functions

	fmt.Printf("\n---------step 2-------------(mark all used declarations)\n\n")

	for _, p := range packages {
		for _, f := range p.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				node, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				currentPos := fileSet.Position(node.Pos())
				currentFile, _ := filepath.Abs(currentPos.Filename)
				// fmt.Println("guru -json definition " + fmt.Sprintf("%s:#%d", fn, pos.Offset))
				out, _ := exec.Command("guru", "-json", "definition", fmt.Sprintf("%s:#%d", currentFile, currentPos.Offset)).Output()
				var def serial.Definition
				json.Unmarshal(out, &def)
				if def.ObjPos == "" {
					return true
				}
				// fmt.Println("found definition of ", node.Name)
				// fmt.Println(def)
				arr := strings.Split(def.ObjPos, ":")
				defLine, _ := strconv.Atoi(arr[1])
				defColumn, _ := strconv.Atoi(arr[2])
				defFile := arr[0] //path.Join(build.Default.GOPATH, "src", w.p.dir, w.f.Name.Name+".go")
				// fmt.Println("found node: ", node.Name)
				// fmt.Println("node at: ", currentFile, currentPos.Line, currentPos.Column)
				// fmt.Println("defn at: ", defFile, defLine, defColumn)
				if currentFile == defFile && currentPos.Line == defLine && currentPos.Column == defColumn {
					return true
				}
				_, ok = declarations[fmt.Sprintf("%s:%d:%d", defFile, defLine, defColumn)]
				if !ok {
					return true
				}
				fmt.Println(node.Name)
				fmt.Println("used in: ", currentFile, currentPos.Line, currentPos.Column)
				fmt.Println("defd in: ", defFile, defLine, defColumn)
				delete(declarations, fmt.Sprintf("%s:%d:%d", defFile, defLine, defColumn))
				return true
			})
		}
	}

	// Step 3: return a list of unused functions

	fmt.Printf("\n---------step 3-------------(list all unused declarations)\n\n")

	for pos, name := range declarations {
		fmt.Printf("* %s = %s\n", name.name, pos)
	}
}
