// Package main содержит multichecker для статического анализа кода.
//
// # Запуск
//
// Для запуска multichecker использовать команду:
//
//	go run ./cmd/staticlint ./...
//
// Для проверки конкретного пакета:
//
//	go run ./cmd/staticlint ./internal/handler/urlhandlers
//
// Для получения справки:
//
//	go run ./cmd/staticlint -help
//
// # Анализаторы
//
// ## Собственный анализатор
//
//   - errexitcheck — запрещает прямой вызов os.Exit в функции main пакета main.
//     Заставляет использовать log.Fatal или возврат ошибки вместо os.Exit.
//
// ## Публичные анализаторы
//
//   - ineffassign — находит неэффективные присваивания.
//     Обнаруживает переменные, которым присваивается значение, но оно не используется.
//
// ## Стандартные анализаторы
//
//   - copylock — проверяет копирование блокировок (sync.Mutex и др.).
//   - loopclosure — проверяет захват переменных цикла в замыканиях.
//   - lostcancel — проверяет отмену контекста.
//   - printf — проверяет формат строк в fmt.Printf и подобных.
//   - structtag — проверяет корректность тегов структур.
//   - unmarshal — проверяет передачу указателей в json.Unmarshal.
//   - unreachable — находит недостижимый код.
//
// ## Анализаторы staticcheck (класс SA)
//
// Все анализаторы класса SA из пакета staticcheck.io.
// Они проверяют:
//   - SA1000-SA1030 — различные ошибки в коде;
//   - обработку ошибок;
//   - работу с типами;
//   - конкурентность;
//   - и другие проблемы.
//
// ## Анализаторы stylecheck (класс ST)
//
//   - ST1001 — проверяет использование dot imports.
//
// ## Анализаторы quickfix (класс QF)
//
//   - QF1002 — предлагает упростить условия.
//   - QF1010 — предлагает использовать sort.Slice.

package main

import (
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	multichecker.Main(
		GetAnalyzers()...,
	)
}

func GetAnalyzers() []*analysis.Analyzer {
	analysers := []*analysis.Analyzer{
		ErrExitAnalyzer,
		ineffassign.Analyzer,
		copylock.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		unreachable.Analyzer,
		unmarshal.Analyzer,
	}

	for _, a := range staticcheck.Analyzers {
		analysers = append(analysers, a.Analyzer)
	}

	checks := map[string]bool{
		"ST1001": true,
		"ST1003": true,
		"QF1002": true,
		"QF1010": true,
	}

	for _, a := range stylecheck.Analyzers {
		if checks[a.Analyzer.Name] {
			analysers = append(analysers, a.Analyzer)
		}
	}
	for _, a := range quickfix.Analyzers {
		if checks[a.Analyzer.Name] {
			analysers = append(analysers, a.Analyzer)
		}
	}

	return analysers
}
