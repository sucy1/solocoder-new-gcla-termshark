// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package ui contains user-interface functions and helpers for termshark.
package ui

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid"
	"github.com/gcla/termshark/v2/configs/profiles"
)

//======================================================================

var HexSearchCaseSensitive bool = false
var HexSearchMatchCount int = 0
var HexSearchTerm string = ""

//======================================================================

func ToggleHexSearchCase(app gowid.IApp) {
	HexSearchCaseSensitive = !HexSearchCaseSensitive
	profiles.SetConf("main.hex-search-case-sensitive", HexSearchCaseSensitive)
	if HexSearchCaseSensitive {
		OpenMessage("Hex search case sensitivity: on", appView, app)
	} else {
		OpenMessage("Hex search case sensitivity: off", appView, app)
	}
}

//======================================================================

func UpdateHexSearchMatchCount(term string, data string, app gowid.IApp) int {
	count := countMatches(term, data, HexSearchCaseSensitive)
	HexSearchMatchCount = count
	HexSearchTerm = term
	return count
}

//======================================================================

func HexSearchStatusText() string {
	if HexSearchTerm != "" {
		caseStr := "insensitive"
		if HexSearchCaseSensitive {
			caseStr = "sensitive"
		}
		return fmt.Sprintf("Search: '%s' | Matches: %d | Case: %s", HexSearchTerm, HexSearchMatchCount, caseStr)
	}
	return ""
}

//======================================================================

func countMatches(term string, data string, caseSensitive bool) int {
	if caseSensitive {
		return strings.Count(data, term)
	}
	return strings.Count(strings.ToLower(data), strings.ToLower(term))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
