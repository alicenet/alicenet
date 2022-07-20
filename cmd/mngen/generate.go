package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func runTemplate(r *registrar, packagestr, tmpl, outfile string) {
	r.Package = packagestr
	output := &strings.Builder{}
	it := template.Must(template.New("").Parse(tmpl))
	err := it.Execute(output, r)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(outfile)
	if err != nil {
		fmt.Printf("File error:%s", err.Error())
		os.Exit(1)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(output.String())
	if err != nil {
		fmt.Printf("File error:%s", err.Error())
		os.Exit(1)
	}
	_ = w.Flush()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func main() {
	infilepath := flag.String("i", "", "Required. The relative path to the file used for input.")
	outpath := flag.String("o", "", "Required. The relative path to the directory used for output.")
	packagestr := flag.String("p", "", "Required. The package name to be used in the generated files.")

	help := flag.Bool("help", false, "Shows help.")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if *infilepath == "" {
		fmt.Println("Missing input file.")
		flag.Usage()
		os.Exit(1)
	}
	if *outpath == "" {
		fmt.Println("Missing output file path.")
		flag.Usage()
		os.Exit(1)
	}
	if *packagestr == "" {
		fmt.Println("Missing the name of the package for output files.")
		flag.Usage()
		os.Exit(1)
	}

	if !fileExists(*infilepath) {
		panic("input file does not exist")
	}
	if !dirExists(*outpath) {
		panic("output path does not exist")
	}
	_, infile := filepath.Split(*infilepath)
	infilename := strings.TrimSuffix(infile, filepath.Ext(infile))
	proto := parseProto(*infilepath)

	runTemplate(proto, *packagestr, implDef, filepath.Join(*outpath, strings.Join([]string{infilename, "mngen.go"}, "_")))
	runTemplate(proto, *packagestr, testDef, filepath.Join(*outpath, strings.Join([]string{infilename, "mngen_test.go"}, "_")))
}
