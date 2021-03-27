package search

import (
	"regexp"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func unionRegexp(values []string) string {
	return "(" + strings.Join(values, ")|(") + ")"
}

func langToFileRegexp(lang string) string {
	lang, _ = enry.GetLanguageByAlias(lang) // Invariant: already validated.
	extensions := enry.GetLanguageExtensions(lang)
	patterns := make([]string, len(extensions))
	for i, e := range extensions {
		// Add `\.ext$` pattern to match files with the given extension.
		patterns[i] = regexp.QuoteMeta(e) + "$"
	}
	return unionRegexp(patterns)
}

func appendMap(values []string, f func(in string) string) []string {
	for _, v := range values {
		values = append(values, f(v))
	}
	return values
}

// Assumes actually Atomic query -> means we need to expand query.Basic to atomic, or assume atomic.
func toTextSearch(q query.Basic) *TextPatternInfo {
	// Handle file: and -file: filters.
	filesInclude, filesExclude := q.IncludeExcludeValues(query.FieldFile)
	// Handle lang: and -lang: filters.
	langInclude, langExclude := q.IncludeExcludeValues(query.FieldLang)
	filesInclude = appendMap(langInclude, langToFileRegexp)
	filesExclude = appendMap(langExclude, langToFileRegexp)
	filesReposMustInclude, filesReposMustExclude := q.IncludeExcludeValues(query.FieldRepoHasFile)
	selector, _ := filter.SelectPathFromString(q.FindValue(query.FieldSelect)) // Invariant: already validated.

	// TODO  handle opts.fileMatchLimit and opts.forceFileSearch (for suggestions)

	return &TextPatternInfo{
		// Atomic Assumptions
		Pattern:         q.Pattern.(query.Pattern).Value,
		IsNegated:       q.Pattern.(query.Pattern).Negated,
		IsRegExp:        q.IsRegexp(),
		IsStructuralPat: q.IsStructural(),

		// Janky -- does this apply on to pattern, or params? Probably only pattern.
		IsCaseSensitive: q.IsCaseSensitive(),

		// Parameters
		IncludePatterns:              filesInclude,
		ExcludePattern:               unionRegexp(filesExclude),
		Languages:                    langInclude,
		FilePatternsReposMustInclude: filesReposMustInclude,
		FilePatternsReposMustExclude: filesReposMustExclude,
		Select:                       selector,
		CombyRule:                    q.FindValue(query.FieldCombyRule),
	}
}
