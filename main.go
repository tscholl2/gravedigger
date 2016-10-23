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

type pkg struct {
	*ast.Package
	declares map[string]*ast.Ident
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

	packages := make(map[string]*pkg)

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
			// get the absoltue path to the directory
			filePathAbs, err := filepath.Abs(filePath)
			if err != nil {
				panic(err)
			}
			packages[filePathAbs] = &pkg{p, make(map[string]*ast.Ident)}
		}
		return nil
	}); err != nil {
		panic(err)
	}

	for p := range packages {
		fmt.Println(p)
	}

	// Step 1: go through and mark all declarations

	fmt.Printf("\n---------step 1-------------(mark all declarations)\n\n")

	for _, pkg := range packages {
		for _, file := range pkg.Files {
			for _, n := range file.Decls {
				switch node := n.(type) {
				case *ast.FuncDecl:
					if node.Name.Name == "init" || node.Name.Name == "_" || (pkg.Name == "main" && node.Name.Name == "main") {
						continue
					}
					pkg.declares[node.Name.Name] = node.Name
				case *ast.GenDecl:
					// var, const, types
					for _, spec := range node.Specs {
						switch s := spec.(type) {
						case *ast.ValueSpec:
							// constants and variables
							for _, n := range s.Names {
								pkg.declares[n.Name] = n
							}
						case *ast.TypeSpec:
							// type definitions
							pkg.declares[s.Name.Name] = s.Name
							// add struct fields?
							/*
								ast.Inspect(s.Type, func(n ast.Node) bool {
									node, ok := n.(*ast.StructType)
									if !ok {
										return true
									}
									for _, node2 := range node.Fields.List {
										for _, node3 := range node2.Names {
											pkg.declares[s.Name.Name+"."+node3.Name] = node3
										}
									}
									return true
								})
							*/
						}
					}
				}
			}
		}
	}

	for k, v := range packages {
		fmt.Printf("%s:\n", k)
		for k2 := range v.declares {
			fmt.Printf("\t- %s\n", k2)
		}
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
				fmt.Println("need to dleete ", node.Name, " from ", defFile)
				packageDir := filepath.Dir(defFile)
				fmt.Println("declaration is in package: ", packageDir)
				delete(packages[packageDir].declares, node.Name)
				return true
			})
		}
	}

	// Step 3: return a list of unused functions

	fmt.Printf("\n---------step 3-------------(list all unused declarations)\n\n")

	unused := make(map[string][]*ast.Ident)
	for _, p := range packages {
		if len(p.declares) == 0 {
			continue
		}
		// fmt.Printf("%s:\n", k)
		for _, node := range p.declares {
			pos := fileSet.Position(node.NamePos)
			filename, _ := filepath.Abs(pos.Filename)
			unused[filename] = append(unused[filename], node)
		}
	}
	for filename, arr := range unused {
		for _, node := range arr {
			pos := fileSet.Position(node.Pos())
			fmt.Printf("%s:%d:%d ---> %s\n", filename, pos.Line, pos.Column, node.Name)
		}
	}
}
