package rules

import (
	"github.com/bmatcuk/doublestar/v4"
	"github.com/driftee-ai/drift/pkg/config"
)

// FilterTriggeredRules filters a list of rules, returning only those that are
// "triggered" by a list of changed files. A rule is triggered if any of the
// changed files match any of its 'code' or 'docs' glob patterns.
// If the changedFiles list is empty, all rules are returned.
func FilterTriggeredRules(rules []config.Rule, changedFiles []string) ([]config.Rule, error) {
	if len(changedFiles) == 0 {
		return rules, nil
	}

	var triggeredRules []config.Rule
	for _, rule := range rules {
		isTriggered := false
		for _, changedFile := range changedFiles {
			// Check against code globs
			for _, glob := range rule.Code {
				if match, err := doublestar.Match(glob, changedFile); err != nil {
					return nil, err
				} else if match {
					isTriggered = true
					break
				}
			}
			if isTriggered {
				break
			}

			// Check against docs globs
			for _, glob := range rule.Docs {
				if match, err := doublestar.Match(glob, changedFile); err != nil {
					return nil, err
				} else if match {
					isTriggered = true
					break
				}
			}
			if isTriggered {
				break
			}
		}

		if isTriggered {
			triggeredRules = append(triggeredRules, rule)
		}
	}

	return triggeredRules, nil
}
