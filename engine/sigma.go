package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sigma "github.com/bradleyjkemp/sigma-go"
	sigeval "github.com/bradleyjkemp/sigma-go/evaluator"
	"github.com/rs/zerolog"
)

type SigmaEngine struct {
	bundle sigeval.RuleEvaluatorBundle
	rules  []sigma.Rule
}

var sigmaEngine *SigmaEngine

func InitSigmaEngine(rulesDir string, logger *zerolog.Logger) error {
	info, err := os.Stat(rulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info().
				Str("rules_dir", rulesDir).
				Msg("no Sigma rules directory found; Sigma engine disabled")
			return nil
		}
		return err
	}
	if !info.IsDir() {
		logger.Warn().
			Str("rules_dir", rulesDir).
			Msg("Sigma rules path is not a directory; Sigma engine disabled")
		return nil
	}

	var rules []sigma.Rule

	err = filepath.WalkDir(rulesDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !isYAML(path) {
			return nil
		}

		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			logger.Warn().
				Err(readErr).
				Str("file", path).
				Msg("failed to read Sigma rule file")
			return nil
		}

		fileType := sigma.InferFileType(contents)
		switch fileType {
		case sigma.RuleFile:
			rule, parseErr := sigma.ParseRule(contents)
			if parseErr != nil {
				logger.Warn().
					Err(parseErr).
					Str("file", path).
					Msg("failed to parse Sigma rule")
				return nil
			}
			rules = append(rules, rule)
			logger.Debug().
				Str("id", rule.ID).
				Str("title", rule.Title).
				Str("file", path).
				Msg("loaded Sigma rule")
		case sigma.ConfigFile:
			logger.Debug().
				Str("file", path).
				Msg("ignoring Sigma config file (not used yet)")
		default:
			logger.Debug().
				Str("file", path).
				Msg("unknown Sigma file type, ignoring")
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(rules) == 0 {
		logger.Info().
			Str("rules_dir", rulesDir).
			Msg("no Sigma rules loaded; Sigma engine disabled")
		return nil
	}

	bundle := sigeval.ForRules(rules)

	sigmaEngine = &SigmaEngine{
		bundle: bundle,
		rules:  rules,
	}

	logger.Info().
		Int("rule_count", len(rules)).
		Msg("Sigma engine initialised")

	return nil
}

func EvaluateSigmaResult(ctx context.Context, r Result, logger *zerolog.Logger) ([]sigeval.RuleResult, error) {
	if sigmaEngine == nil {
		logger.Debug().Msg("Sigma engine not initialised or no rules – skipping evaluation")
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	event := resultToSigmaEvent(r, logger)

	logger.Debug().
		Interface("sigma_event", event).
		Msg("Sigma event (flattened) generated from Result")

	if flat, ok := event.(map[string]any); ok {
		if title, ok := flat["result.head.title"]; ok {
			logger.Debug().
				Interface("result.head.title", title).
				Msg("Sigma event result.head.title field")
		}
	}

	matches, err := sigmaEngine.bundle.Matches(ctx, event)
	if err != nil {
		logger.Warn().
			Err(err).
			Msg("Sigma evaluation error")
		return nil, err
	}

	if len(matches) == 0 {
		logger.Debug().Msg("Sigma evaluation produced zero RuleResults")
		return nil, nil
	}

	for _, m := range matches {
		logger.Debug().
			Str("rule_id", m.Rule.ID).
			Str("rule_title", m.Rule.Title).
			Bool("match", m.Match).
			Interface("search_results", m.SearchResults).
			Interface("condition_results", m.ConditionResults).
			Msg("Sigma rule evaluation result")

		if m.Match {
			logger.Warn().
				Str("rule_id", m.Rule.ID).
				Str("rule_title", m.Rule.Title).
				Msg("Sigma rule matched scanner result")
		}
	}

	return matches, nil
}

func resultToSigmaEvent(r Result, logger *zerolog.Logger) sigeval.Event {
	b, err := json.Marshal(r)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to marshal Result into JSON for Sigma event")
		return map[string]any{}
	}

	logger.Debug().
		RawJSON("result_json", b).
		Msg("raw Result JSON before Sigma event mapping")

	var nested map[string]any
	if err := json.Unmarshal(b, &nested); err != nil {
		logger.Warn().Err(err).Msg("failed to unmarshal Result JSON into nested map")
		return map[string]any{}
	}

	flat := make(map[string]any)
	flattenMap("", nested, flat)
	return flat
}

func flattenMap(prefix string, v any, out map[string]any) {
	switch val := v.(type) {
	case map[string]any:
		for k, inner := range val {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenMap(key, inner, out)
		}
	case []any:
		for i, inner := range val {
			key := prefix + "[" + fmt.Sprint(i) + "]"
			flattenMap(key, inner, out)
		}
	default:
		if prefix != "" {
			out[prefix] = val
		}
	}
}

func isYAML(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".yaml")
}
