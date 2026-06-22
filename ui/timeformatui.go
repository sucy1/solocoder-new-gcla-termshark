// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package ui

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/termshark/v2/configs/profiles"
)

//======================================================================

type TimeFormatMode int

const (
	TimeFormatDefault TimeFormatMode = iota
	TimeFormatRelative
	TimeFormatDelta
	TimeFormatEpoch
	TimeFormatCount
)

var CurrentTimeFormat TimeFormatMode = TimeFormatDefault

//======================================================================

func CycleTimeFormat(app gowid.IApp) {
	CurrentTimeFormat = (CurrentTimeFormat + 1) % TimeFormatCount

	flag := timeFormatFlag()
	psmlArgs := profiles.ConfStringSlice("main.psml-args", []string{})

	var updated []string
	for i := 0; i < len(psmlArgs); i++ {
		if psmlArgs[i] == "-t" {
			i++
			continue
		}
		updated = append(updated, psmlArgs[i])
	}

	if flag != "" {
		updated = append(updated, "-t", flag)
	}

	profiles.SetConf("main.psml-args", updated)

	RequestReload(app)

	OpenMessage(fmt.Sprintf("Time format: %s", timeFormatName()), appView, app)
}

//======================================================================

func timeFormatFlag() string {
	switch CurrentTimeFormat {
	case TimeFormatDefault:
		return ""
	case TimeFormatRelative:
		return "r"
	case TimeFormatDelta:
		return "d"
	case TimeFormatEpoch:
		return "e"
	default:
		return ""
	}
}

//======================================================================

func timeFormatName() string {
	switch CurrentTimeFormat {
	case TimeFormatDefault:
		return "Default"
	case TimeFormatRelative:
		return "Relative"
	case TimeFormatDelta:
		return "Delta"
	case TimeFormatEpoch:
		return "Epoch seconds"
	default:
		return "Default"
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
