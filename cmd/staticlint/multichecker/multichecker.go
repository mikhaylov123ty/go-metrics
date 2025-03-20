// Модуль multichecker производит статический анализ кода
package multichecker

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"mycheck/analyzer/osexit"
)

// configPath - задает путь до файла конфигурации из корневой папки
// переделать на флаг
const configPath = "./cmd/staticlint/config.json"

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Staticcheck []string `json:"staticChecks"`
}

// Run - запускает статический анализ кода проекта
func Run() error {
	cfgData, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg ConfigData
	if err = json.Unmarshal(cfgData, &cfg); err != nil {
		return err
	}

	mychecks := cfg.parseChecks()

	mychecks = append(mychecks,
		loopclosure.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		errcheck.Analyzer,
		osexit.Analyzer,
	)

	multichecker.Main(
		mychecks...,
	)
	return nil
}

// parseChecks метод парсинга конфигурации инструментов анализа
func (c *ConfigData) parseChecks() []*analysis.Analyzer {
	var res []*analysis.Analyzer
	for _, check := range c.Staticcheck {
		switch {
		case strings.Contains(check, "SA"):
			if check == "SA*" {
				for _, a := range staticcheck.Analyzers {
					res = append(res, a.Analyzer)
				}
				continue
			}
			for _, a := range staticcheck.Analyzers {
				if check == a.Analyzer.Name {
					res = append(res, a.Analyzer)
				}
			}
		case strings.Contains(check, "S1"):
			if check == "S1*" {
				for _, a := range simple.Analyzers {
					res = append(res, a.Analyzer)
				}
				continue
			}
			for _, a := range simple.Analyzers {
				if check == a.Analyzer.Name {
					res = append(res, a.Analyzer)
				}
			}
		case strings.Contains(check, "ST"):
			if check == "ST*" {
				for _, a := range stylecheck.Analyzers {
					res = append(res, a.Analyzer)
				}
				continue
			}
			for _, a := range stylecheck.Analyzers {
				if check == a.Analyzer.Name {
					res = append(res, a.Analyzer)
				}
			}
		case strings.Contains(check, "QF"):
			if check == "QF*" {
				for _, a := range quickfix.Analyzers {
					res = append(res, a.Analyzer)
				}
				continue
			}
			for _, a := range quickfix.Analyzers {
				if check == a.Analyzer.Name {
					res = append(res, a.Analyzer)
				}
			}
		}
	}

	return res
}
