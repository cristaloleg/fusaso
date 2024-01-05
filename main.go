package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/cristaloleg/cnf"
)

func main() {
	fset := flag.NewFlagSet("fusaso", flag.ContinueOnError)

	filename := fset.String("problem", "problem.cnf", "file with the problem (DIMACS format)")
	output := fset.String("out", "fuzz_solve_test.go", "file with the codegen")
	run := fset.Bool("run", false, "run the fuzzing")

	if err := fset.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	problemFile, err := os.Open(*filename)
	if err != nil {
		panic(err)
	}
	defer problemFile.Close()

	outfile, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	problem, err := cnf.ParseDIMAC(problemFile)
	if err != nil {
		panic(err)
	}

	if err := codegen(outfile, problem); err != nil {
		panic(err)
	}

	if *run {
		cmd := exec.Command("go", "test", "-run=^$", "-fuzz=FuzzSolve", *output)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

func codegen(w io.Writer, p *cnf.Problem) error {
	p.Formula.SortBySize()

	data := struct {
		Variables int
		Cases     []string
	}{
		Variables: p.Variables,
		Cases:     toCases(p.Formula),
	}

	return template.Must(template.New("").Parse(codeTmpl)).Execute(w, data)
}

//go:embed codegen.gotmpl
var codeTmpl string

func toCases(formula cnf.Formula) []string {
	res := make([]string, len(formula))
	for i, c := range formula {
		res[i] = conv(c)
	}
	return res
}

func conv(clause cnf.Clause) string {
	var b strings.Builder
	for i, lit := range clause {
		if i > 0 {
			fmt.Fprint(&b, " && ")
		}
		if lit.Sign() {
			fmt.Fprintf(&b, `x[%d] == '0'`, lit.Var()-1)
		} else {
			fmt.Fprintf(&b, `x[%d] == '1'`, lit.Var()-1)
		}
	}
	return b.String()
}
