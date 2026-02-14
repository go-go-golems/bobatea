package repl

import "time"

type helpDrawerModel struct {
	provider HelpDrawerProvider

	visible bool
	doc     HelpDrawerDocument

	reqSeq     uint64
	debounce   time.Duration
	reqTimeout time.Duration

	loading bool
	err     error
	pinned  bool

	prefetch      bool
	dock          HelpDrawerDock
	widthPercent  int
	heightPercent int
	margin        int
}
