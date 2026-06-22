// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package ui contains user-interface functions and helpers for termshark.
package ui

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/dialog"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/termshark/v2/pkg/annotations"
)

//======================================================================

var AnnotationStore *annotations.AnnotationStore

//======================================================================

func InitAnnotationStore() error {
	store, err := annotations.NewStore()
	if err != nil {
		return err
	}
	AnnotationStore = store
	return nil
}

//======================================================================

func getAnnotationForPacket(packetNum int) string {
	if AnnotationStore == nil {
		return ""
	}
	a, ok := AnnotationStore.Get(packetNum)
	if !ok {
		return ""
	}
	return a.Text
}

//======================================================================

func openAnnotationDialog(app gowid.IApp) {
	if AnnotationStore == nil {
		OpenError("Annotation store not initialized.", app)
		return
	}

	p, err := packetNumberFromCurrentTableRow()
	if err != nil {
		OpenError(fmt.Sprintf("Could not determine packet number: %v", err), app)
		return
	}

	packetNum := p.Pos
	existing := getAnnotationForPacket(packetNum)

	editWidget := edit.New()
	if existing != "" {
		editWidget.SetText(existing, app)
	}

	var annDialog *dialog.Widget

	saveBtn := dialog.Button{
		Msg: "Save",
		Action: gowid.MakeWidgetCallback("cb", gowid.WidgetChangedFunction(func(app gowid.IApp, _ gowid.IWidget) {
			annDialog.Close(app)
			txt := editWidget.Text()
			if txt == "" {
				return
			}
			AnnotationStore.Set(packetNum, txt)
			err := AnnotationStore.Save()
			if err != nil {
				OpenError(fmt.Sprintf("Could not save annotation: %v", err), app)
			}
		})),
	}

	deleteBtn := dialog.Button{
		Msg: "Delete",
		Action: gowid.MakeWidgetCallback("cb", gowid.WidgetChangedFunction(func(app gowid.IApp, _ gowid.IWidget) {
			annDialog.Close(app)
			AnnotationStore.Delete(packetNum)
			err := AnnotationStore.Save()
			if err != nil {
				OpenError(fmt.Sprintf("Could not save annotation: %v", err), app)
			}
		})),
	}

	label := fmt.Sprintf("Annotation for packet %d:", packetNum)

	dialogWidgets := make([]interface{}, 0, 4)
	dialogWidgets = append(dialogWidgets,
		text.New(label),
		divider.NewBlank(),
		framed.NewUnicode(editWidget),
	)

	body := framed.NewSpace(pile.NewFlow(dialogWidgets...))

	annDialog = dialog.New(
		body,
		dialog.Options{
			Buttons:         []dialog.Button{saveBtn, deleteBtn, dialog.Cancel},
			NoShadow:        true,
			BackgroundStyle: gowid.MakePaletteRef("dialog"),
			BorderStyle:     gowid.MakePaletteRef("dialog"),
			ButtonStyle:     gowid.MakePaletteRef("dialog-button"),
			Modal:           true,
			FocusOnWidget:   true,
		},
	)

	annDialog.Open(appView, ratio(0.5), app)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
