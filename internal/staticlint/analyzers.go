package staticlint

import (
	"golang.org/x/tools/go/analysis"

	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/tdakkota/asciicheck"

	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func GetAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer
	analyzers = append(analyzers, getStandardAnalyzers()...)
	analyzers = append(analyzers, getCustomAnalyzers()...)
	analyzers = append(analyzers, getStaticCheckAnalyzers()...)
	analyzers = append(analyzers, getPublicAnalyzers()...)
	return analyzers
}

func getStandardAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)

	return analyzers
}

func getCustomAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers, NoOsExitAnalyzer)

	return analyzers
}

func getStaticCheckAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer

	blacklist := map[string]struct{}{
		"ST1000": {}, // Package comment
	}

	for _, lintAnalyzer := range simple.Analyzers {
		if _, blacklisted := blacklist[lintAnalyzer.Analyzer.Name]; !blacklisted {
			analyzers = append(analyzers, lintAnalyzer.Analyzer)
		}
	}

	for _, lintAnalyzer := range staticcheck.Analyzers {
		if _, blacklisted := blacklist[lintAnalyzer.Analyzer.Name]; !blacklisted {
			analyzers = append(analyzers, lintAnalyzer.Analyzer)
		}
	}

	for _, lintAnalyzer := range stylecheck.Analyzers {
		if _, blacklisted := blacklist[lintAnalyzer.Analyzer.Name]; !blacklisted {
			analyzers = append(analyzers, lintAnalyzer.Analyzer)
		}
	}

	for _, lintAnalyzer := range quickfix.Analyzers {
		if _, blacklisted := blacklist[lintAnalyzer.Analyzer.Name]; !blacklisted {
			analyzers = append(analyzers, lintAnalyzer.Analyzer)
		}
	}

	return analyzers
}

func getPublicAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		errcheck.Analyzer,
		asciicheck.NewAnalyzer(),
	}
}
