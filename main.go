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
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/oracle/serial"
)

var (
	dir = "."
)

type decl struct {
	f *ast.File
	n *ast.Ident
}

type pkg struct {
	*ast.Package
	dir      string
	declares map[string]*decl
}

func main() {
	flag.CommandLine.Usage = func() {
		fmt.Println(`gravedigger: looks for unused code in a directory. This differs
from other packages in that it takes a directory and lists things that are unused
any where in that directory, including exported things in subpackages/subdirectories.
Example: 'gravedigger'
Example: 'gravedigger .'
Example: 'gravedigger cmd/'
Options:`)
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

	fmt.Printf("\n---------step 0-------------(find all code in directory)\n\n")

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
			// get the directory of the current directory relative to the GOPATH
			filePathAbs, err := filepath.Abs(filePath)
			if err != nil {
				panic(err)
			}
			filePathRelGoPath, err := filepath.Rel(path.Join(build.Default.GOPATH, "src"), filePathAbs)
			if err != nil {
				panic(err)
			}
			packages[filePathRelGoPath] = &pkg{p, filePathRelGoPath, make(map[string]*decl)}
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
					pkg.declares[node.Name.Name] = &decl{file, node.Name}
				case *ast.GenDecl:
					// var, const, types
					for _, spec := range node.Specs {
						switch s := spec.(type) {
						case *ast.ValueSpec:
							// constants and variables.
							for _, n := range s.Names {
								pkg.declares[n.Name] = &decl{file, n}
							}
						case *ast.TypeSpec:
							// type definitions.
							pkg.declares[s.Name.Name] = &decl{file, s.Name}
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

	type fp struct {
		f *ast.File
		p *pkg
	}
	var files []fp
	for _, p := range packages {
		for _, f := range p.Files {
			files = append(files, fp{f, p})
		}
	}
	for _, fp := range files {
		ast.Walk(&walker{fp.p, fp.f, packages, fileSet}, fp.f)
	}

	// Step 3: return a list of unused functions

	fmt.Printf("\n---------step 3-------------(list all unused declarations)\n\n")

	for k, v := range packages {
		if len(v.declares) == 0 {
			continue
		}
		fmt.Printf("%s:\n", k)
		for k2, v2 := range v.declares {
			// TODO get line number
			fmt.Printf("\t- %s ---> %s.go:%s\n", k2, v2.f.Name.Name, fileSet.Position(v2.n.Pos()))
		}
	}

}

type walker struct {
	p  *pkg
	f  *ast.File
	ps map[string]*pkg
	fs *token.FileSet
}

func (w *walker) Visit(n ast.Node) ast.Visitor {
	switch node := n.(type) {
	case *ast.AssignStmt:
		// go through RHS of assignment
		for _, v := range node.Rhs {
			ast.Walk(w, v)
		}
	case *ast.ValueSpec:
		// go through variable initializers
		for _, v := range node.Values {
			ast.Walk(w, v)
		}
		if node.Type != nil {
			ast.Walk(w, node.Type)
		}
	case *ast.BlockStmt:
		// body of statement
		for _, stmt := range node.List {
			ast.Walk(w, stmt)
		}
	case *ast.FuncDecl:
		// function signatures
		ast.Walk(w, node.Type)
	case *ast.TypeSpec:
		// type definitions
		ast.Walk(w, node.Type)
	case *ast.Ident:
		currentPos := w.fs.Position(node.Pos())
		currentFile, _ := filepath.Abs(currentPos.Filename)
		// fmt.Println("guru -json definition " + fmt.Sprintf("%s:#%d", fn, pos.Offset))
		out, _ := exec.Command("guru", "-json", "definition", fmt.Sprintf("%s:#%d", currentFile, currentPos.Offset)).Output()
		var def serial.Definition
		json.Unmarshal(out, &def)
		if def.ObjPos == "" {
			return w
		}
		// fmt.Println("found definition of ", node.Name)
		// fmt.Println(def)
		arr := strings.Split(def.ObjPos, ":")
		defLine, _ := strconv.Atoi(arr[1])
		defColumn, _ := strconv.Atoi(arr[2])
		defFile := arr[0] //path.Join(build.Default.GOPATH, "src", w.p.dir, w.f.Name.Name+".go")
		// fmt.Println("found node: ", node.Name)
		fmt.Println("node at: ", currentFile, currentPos.Line, currentPos.Column)
		fmt.Println("defn at: ", defFile, defLine, defColumn)

		// if filename == currentFile && pos.Line == lineNumber && pos.Column == columnNumber {
		//	return w
		// }

	}
	return w
}
