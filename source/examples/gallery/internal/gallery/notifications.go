package gallery

import (
	"errors"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

const notificationsPage galleryPage = shellPage + 1

type notificationDemo struct {
	current *mygo.Notification
	pending bool
	silent  bool
	// sender sends the notification as a message from the author, with
	// their avatar where the system shows one.
	sender bool
	// picture shows the author's avatar with the notification's text.
	picture bool
	status  string
	err     error
}

func (d *notificationDemo) close() {
	if d.current != nil {
		d.current.Close()
	}
	d.current = nil
	d.pending = false
}

func (g *Gallery) sendNotification() {
	d := &g.notification
	if g.Shell == nil || d.pending {
		return
	}
	d.err = nil
	if !mygo.NotificationsSupported() {
		d.status = "notificationUnsupported"
		return
	}
	d.close()
	o := mygo.NotificationOptions{
		Title:  g.text("notificationTitle"),
		Body:   g.text("notificationBody"),
		Silent: d.silent,
	}
	if avatar, err := galleryImageData(); d.sender && err == nil {
		// macOS shows the avatar only in an app signed with the Communication
		// Notifications capability; elsewhere the notification is ordinary.
		o.Sender = &mygo.NotificationSender{ID: "author", Name: "ONEO", Avatar: avatar}
	}
	if picture, err := galleryImageData(); d.picture && err == nil {
		o.Image = picture
	}
	n := mygo.NewNotification(o)
	d.current, d.pending = n, true
	d.status = "notificationSending"
	n.OnClick(func() {
		if d.current != n {
			return
		}
		d.close()
		d.status = "notificationClicked"
		g.page = notificationsPage
		g.settings.open = false
		g.showMainWindow()
		g.Shell.Window().Invalidate()
	})
	// Show may wait for the macOS permission prompt. Keep the UI responsive;
	// all gallery state and completion handling remain on the UI thread.
	go func() {
		err := n.Show()
		mygo.RunOnMain(func() {
			if d.current != n {
				// Withdrawal or Gallery shutdown may precede delivery.
				n.Close()
				return
			}
			d.pending = false
			switch {
			case errors.Is(err, mygo.ErrNotificationsDenied):
				d.close()
				d.status = "notificationDenied"
			case err != nil:
				d.close()
				d.status, d.err = "notificationFailed", err
			default:
				d.status = "notificationSent"
			}
			g.Shell.Window().Invalidate()
		})
	}()
}

func (g *Gallery) Notifications(c *ui.Context) {
	g.withLanguage(c, func() {
		d := &g.notification
		components.ScrollView(c, components.ScrollViewOptions{}).Fill().Children(func() {
			ui.Column(c).FillWidth().Padding(c.Theme().Space(6)).Gap(16).Children(func() {
				components.Checkbox(c, &d.silent, tr(c, "notificationSilent"))
				components.Checkbox(c, &d.sender, tr(c, "notificationSender"))
				components.Checkbox(c, &d.picture, tr(c, "notificationPicture"))
				components.Button(c, tr(c, "notificationSend"), components.ButtonPrimary).
					Disabled(g.Shell == nil || d.pending).OnClick(g.sendNotification)
				components.Button(c, tr(c, "notificationWithdraw"), components.ButtonOutline).
					Disabled(d.current == nil).OnClick(func() {
					d.close()
					d.err = nil
					d.status = "notificationWithdrawn"
				})
				if d.status != "" {
					components.Text(c, tr(c, d.status)).Wrap()
				}
				if d.err != nil {
					components.Text(c, d.err.Error()).Wrap().TextColor(c.Theme().Danger)
				}
			})
		})
	})
}
