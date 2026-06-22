// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package ui contains user-interface functions and helpers for termshark.
package ui

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid"
	"github.com/gcla/termshark/v2/pkg/presets"
	"github.com/gcla/termshark/v2/widgets/minibuffer"
	"github.com/shibukawa/configdir"
)

//======================================================================

var PresetStore *presets.PresetStore

func InitPresetStore() {
	dir := configdir.New("", "termshark").QueryFolders(configdir.Global)[0].Path
	PresetStore = presets.NewStore(dir)
}

//======================================================================

type presetCommand struct{}

var _ minibuffer.IAction = presetCommand{}

func (d presetCommand) Run(app gowid.IApp, args ...string) error {
	var err error

	if len(args) < 2 {
		err = fmt.Errorf("Invalid preset command")
	} else {
		switch args[1] {
		case "save":
			if len(args) < 3 {
				err = fmt.Errorf("Usage: preset save NAME")
			} else {
				saveCurrentFilterAsPreset(args[2], app)
			}
		case "load":
			if len(args) < 3 {
				err = fmt.Errorf("Usage: preset load NAME")
			} else {
				applyPreset(args[2], app)
			}
		case "delete":
			if len(args) < 3 {
				err = fmt.Errorf("Usage: preset delete NAME")
			} else {
				err = PresetStore.Delete(args[2])
				if err == nil {
					OpenMessage(fmt.Sprintf("Preset %s deleted.", args[2]), appView, app)
				}
			}
		case "list":
			presetsList := PresetStore.List()
			if len(presetsList) == 0 {
				OpenMessage("No presets defined.", appView, app)
			} else {
				lines := make([]string, len(presetsList))
				for i, p := range presetsList {
					lines[i] = fmt.Sprintf("%s: %s", p.Name, p.Filter)
				}
				OpenMessage(strings.Join(lines, "\n"), appView, app)
			}
		default:
			err = fmt.Errorf("Invalid preset command: %s", args[1])
		}
	}

	if err != nil {
		OpenMessage(fmt.Sprintf("Error: %s", err), appView, app)
	}

	return err
}

func (d presetCommand) OfferCompletion() bool {
	return true
}

func (d presetCommand) Arguments(toks []string, app gowid.IApp) []minibuffer.IArg {
	res := make([]minibuffer.IArg, 0)
	pref := ""
	if len(toks) > 0 {
		pref = toks[0]
	}
	res = append(res, presetArg{substr: pref})
	return res
}

//======================================================================

type presetArg struct {
	substr string
}

var _ minibuffer.IArg = presetArg{}

func (s presetArg) OfferCompletion() bool {
	return true
}

func (s presetArg) Completions() []string {
	matches := make([]string, 0)
	names := PresetStore.Names()
	for _, n := range names {
		if strings.Contains(n, s.substr) {
			matches = append(matches, n)
		}
	}
	return matches
}

//======================================================================

func applyPreset(name string, app gowid.IApp) {
	p, ok := PresetStore.Get(name)
	if !ok {
		OpenMessage(fmt.Sprintf("Preset %s not found.", name), appView, app)
		return
	}
	setFocusOnDisplayFilter(app)
	FilterWidget.SetValue(p.Filter, app)
}

//======================================================================

func saveCurrentFilterAsPreset(name string, app gowid.IApp) {
	filter := FilterWidget.Value()
	err := PresetStore.Save(name, filter)
	if err != nil {
		OpenMessage(fmt.Sprintf("Error saving preset: %s", err), appView, app)
		return
	}
	OpenMessage(fmt.Sprintf("Preset %s saved.", name), appView, app)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
