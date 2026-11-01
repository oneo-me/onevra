package gallery

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/examples/gallery/internal/gallery/locales"
)

// dialogTypes follow the order of the message type tabs.
var dialogTypes = []mygo.MessageType{mygo.MessageInfo, mygo.MessageWarning, mygo.MessageError, mygo.MessageQuestion}

// dialogParent attaches dialogs to the gallery window, as sheets on macOS,
// unless the demonstration detaches them.
func (d *systemDemo) dialogParent() *mygo.Window {
	if d.dialogDetached {
		return nil
	}
	return d.window()
}

// showDialog runs a blocking dialog off the UI thread and shows its outcome
// as label and value rows, none when the dialog was canceled.
func (d *systemDemo) showDialog(m locales.Messages, show func() ([][2]string, error)) {
	d.background(func() func() {
		rows, err := show()
		return func() {
			d.dialogRows, d.dialogNote = rows, ""
			switch {
			case err != nil:
				d.dialogNote = m.SystemError(err.Error())
			case len(rows) == 0:
				d.dialogNote = m.DialogCanceled()
			}
		}
	})
}

// pathRows labels each chosen path.
func pathRows(m locales.Messages, paths []string, err error) ([][2]string, error) {
	rows := make([][2]string, len(paths))
	for i, path := range paths {
		rows[i] = [2]string{m.DialogPath(), path}
	}
	return rows, err
}

func exampleNativeDialogs(c *ui.Context, s *demoState) {
	d, m, t := &s.system, msg(c), c.Theme()
	ui.Row(c).Gap(8).Wrap().Children(func() {
		if components.Button(c, m.DialogOpenFiles(), components.ButtonPrimary).Clicked() {
			d.showDialog(m, func() ([][2]string, error) {
				paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
					Parent:   d.dialogParent(),
					Title:    m.DialogOpenFiles(),
					Filters:  []mygo.FileFilter{{Name: m.DialogImages(), Extensions: []string{"png", "jpg", "svg"}}},
					Multiple: true,
				})
				return pathRows(m, paths, err)
			})
		}
		if components.Button(c, m.DialogOpenFolder(), components.ButtonOutline).Clicked() {
			d.showDialog(m, func() ([][2]string, error) {
				paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
					Parent:            d.dialogParent(),
					Title:             m.DialogOpenFolder(),
					Directory:         true,
					CreateDirectories: true,
				})
				return pathRows(m, paths, err)
			})
		}
		if components.Button(c, m.DialogSave(), components.ButtonOutline).Clicked() {
			d.showDialog(m, func() ([][2]string, error) {
				path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
					Parent:            d.dialogParent(),
					Title:             m.DialogSave(),
					DefaultPath:       "notes.json",
					Filters:           []mygo.FileFilter{{Name: "JSON", Extensions: []string{"json"}}},
					CreateDirectories: true,
				})
				if path == "" {
					return nil, err
				}
				return pathRows(m, []string{path}, err)
			})
		}
	})
	components.Checkbox(c, &d.dialogDetached, m.DialogDetached())

	components.Tabs(c, &d.dialogType, components.TabsSegmented,
		m.DialogInfo(), m.DialogWarning(), m.DialogErrorType(), m.DialogQuestion())
	ui.Row(c).Gap(8).Wrap().Children(func() {
		if components.Button(c, m.DialogAsk(), components.ButtonPrimary).Clicked() {
			kind := dialogTypes[d.dialogType]
			buttons := []string{m.Delete(), m.Cancel()}
			d.showDialog(m, func() ([][2]string, error) {
				res, err := mygo.Dialog.Message(mygo.MessageOptions{
					Parent:  d.dialogParent(),
					Type:    kind,
					Message: m.DialogAskMessage(),
					Detail:  m.DialogAskDetail(),
					Buttons: buttons,
					// Delete erases data, so macOS and Linux set it apart.
					DestructiveButtons: []int{0},
					CheckboxLabel:      m.DialogDontAsk(),
				})
				if err != nil {
					return nil, err
				}
				return [][2]string{{m.DialogButton(), buttons[res.Button]}, {m.DialogDontAsk(), yesNo(m, res.CheckboxChecked)}}, nil
			})
		}
		if components.Button(c, m.DialogShowError(), components.ButtonOutline).Clicked() {
			d.background(func() func() {
				mygo.Dialog.Error(m.DialogErrorTitle(), m.DialogErrorDetail())
				return func() {}
			})
		}
	})

	switch {
	case d.dialogNote != "":
		components.Text(c, d.dialogNote).Wrap().TextColor(t.TextMuted)
	case len(d.dialogRows) == 0:
		components.Text(c, m.DialogEmpty()).TextColor(t.TextMuted)
	}
	for i, row := range d.dialogRows {
		ui.Column(c).Key(i).FillWidth().Children(func() { systemRow(c, row[0], row[1]) })
	}
}
