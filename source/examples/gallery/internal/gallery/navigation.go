package gallery

// Component sections describe purpose, never implementation ownership.
var navigationGroups = []string{
	"Shaders", "Conversation", "Plugins", "Actions & navigation", "Text input", "Selection & values",
	"Date & time", "Color", "Content display", "Code & diffs", "Charts", "Layout & containers",
	"Feedback & overlays", "Games",
}

// newPages and updatedPages hold only the routes added or changed since the
// latest published version (the newest source/v* tag); the top New components
// and Updated components groups list them in place of their purpose groups.
// Publishing leaves both lists unchanged, so the release still marks what it
// ships. The first change after a release removes every route already listed
// at that tag, unless the route changed again since, before recording its
// own. A route is in at most one list; a route added since the release stays
// new however often it changes.
var newPages = map[string]bool{
	"mygo/Strokes": true, "mygo/Hidden elements": true, "mygo/Zoom": true,
}

var updatedPages = map[string]bool{
	"mygo/WebView": true, "mygo/Glass": true, "mygo/Audio": true, "mygo/Flyout": true, "mygo/Markdown": true,
	"mygo/Code editor": true, "mygo/System information": true, "mygo/Motion": true,
}

type navigationEntry struct {
	label string
	page  galleryPage
	// mygo marks the single sidebar entry that hosts every MyGo page behind
	// its own in-page list; page is then only the first of them.
	mygo bool
}

// mygoGroupLabel names the sidebar group of the MyGo entry, and the command
// palette group of the pages it hosts.
const mygoGroupLabel = "MyGo"

// newGroupLabel names the group of the pages in newPages, and
// updatedGroupLabel the group of the pages in updatedPages.
const (
	newGroupLabel     = "New components"
	updatedGroupLabel = "Updated components"
)

type navigationGroup struct {
	label   string
	entries []navigationEntry
}

func navigationCatalog() []navigationGroup {
	var groups []navigationGroup
	add := func(label string, entries []navigationEntry) {
		if len(entries) > 0 {
			groups = append(groups, navigationGroup{label: label, entries: entries})
		}
	}
	add("Overview", []navigationEntry{{"Welcome", welcomePage, false}})
	add(newGroupLabel, newNavigationEntries(false))
	add(updatedGroupLabel, updatedNavigationEntries(false))
	add("Icon library", []navigationEntry{{"Lucide", lucidePage, false}})
	add(mygoGroupLabel, []navigationEntry{{"Official components", notificationsPage, true}})
	for _, group := range navigationGroups {
		var entries []navigationEntry
		for i, example := range controlCatalog {
			if example.group == group && !newPages["controls/"+example.name] && !updatedPages["controls/"+example.name] {
				entries = append(entries, navigationEntry{example.label, controlPageStart() + galleryPage(i), false})
			}
		}
		add(group, entries)
	}
	return groups
}

// mygoNavigationGroups lists the pages of the MyGo entry for its in-page
// list: its own New and Updated components groups, then purposes in the
// sidebar's order.
func mygoNavigationGroups() []navigationGroup {
	var groups []navigationGroup
	add := func(label string, entries []navigationEntry) {
		if len(entries) > 0 {
			groups = append(groups, navigationGroup{label: label, entries: entries})
		}
	}
	add(newGroupLabel, newNavigationEntries(true))
	add(updatedGroupLabel, updatedNavigationEntries(true))
	for _, group := range append([]string{"Desktop"}, navigationGroups...) {
		var entries []navigationEntry
		if group == "Desktop" {
			entries = []navigationEntry{{"System notifications", notificationsPage, false}}
		}
		for i, example := range mygoCatalog {
			if example.group == group && !newPages["mygo/"+example.name] && !updatedPages["mygo/"+example.name] {
				entries = append(entries, navigationEntry{example.name, mygoPageStart + galleryPage(i), false})
			}
		}
		if len(entries) > 0 {
			groups = append(groups, navigationGroup{label: group, entries: entries})
		}
	}
	return groups
}

// isMyGoPage reports pages behind the MyGo entry: System notifications and
// the MyGo catalog, which directly follow one another.
func isMyGoPage(page galleryPage) bool {
	return page >= notificationsPage && page < controlPageStart()
}

// newNavigationEntries collects the pages in newPages in catalog order, either
// Onevra's for the sidebar or MyGo's for the MyGo entry's in-page list.
func newNavigationEntries(mygo bool) []navigationEntry {
	return navigationEntries(newPages, mygo)
}

// updatedNavigationEntries collects the pages in updatedPages in catalog order.
func updatedNavigationEntries(mygo bool) []navigationEntry {
	return navigationEntries(updatedPages, mygo)
}

func navigationEntries(pages map[string]bool, mygo bool) []navigationEntry {
	var entries []navigationEntry
	if !mygo {
		for i, example := range controlCatalog {
			if pages["controls/"+example.name] {
				entries = append(entries, navigationEntry{example.label, controlPageStart() + galleryPage(i), false})
			}
		}
		return entries
	}
	for i, example := range mygoCatalog {
		if pages["mygo/"+example.name] {
			entries = append(entries, navigationEntry{example.name, mygoPageStart + galleryPage(i), false})
		}
	}
	return entries
}
