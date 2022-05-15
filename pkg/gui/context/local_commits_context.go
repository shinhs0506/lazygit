package context

import (
	"github.com/jesseduffield/generics/set"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type LocalCommitsContext struct {
	*LocalCommitsViewModel
	*ViewportListContextTrait
}

var _ types.IListContext = (*LocalCommitsContext)(nil)

func NewLocalCommitsContext(
	getModel func() []*models.Commit,
	guiContextState GuiContextState,
	view *gocui.View,

	onFocus func(...types.OnFocusOpts) error,
	onRenderToMain func(...types.OnFocusOpts) error,
	onFocusLost func() error,

	c *types.HelperCommon,
) *LocalCommitsContext {
	viewModel := NewLocalCommitsViewModel(getModel, c, guiContextState.Needle)

	getDisplayStrings := func(startIdx int, length int) [][]string {
		selectedCommitSha := ""
		if guiContextState.IsFocused() {
			selectedCommit := viewModel.GetSelected()
			if selectedCommit != nil {
				selectedCommitSha = selectedCommit.Sha
			}
		}
		return presentation.GetCommitListDisplayStrings(
			viewModel.getModel(),
			guiContextState.ScreenMode() != types.SCREEN_NORMAL,
			cherryPickedCommitShaSet(guiContextState),
			guiContextState.Modes().Diffing.Ref,
			c.UserConfig.Gui.TimeFormat,
			c.UserConfig.Git.ParseEmoji,
			selectedCommitSha,
			startIdx,
			length,
			shouldShowGraph(guiContextState, c.UserConfig),
			guiContextState.BisectInfo(),
		)
	}

	return &LocalCommitsContext{
		LocalCommitsViewModel: viewModel,
		ViewportListContextTrait: &ViewportListContextTrait{
			ListContextTrait: &ListContextTrait{
				Context: NewSimpleContext(NewBaseContext(NewBaseContextOpts{
					ViewName:   "commits",
					WindowName: "commits",
					Key:        LOCAL_COMMITS_CONTEXT_KEY,
					Kind:       types.SIDE_CONTEXT,
					Focusable:  true,
				}), ContextCallbackOpts{
					OnFocus:        onFocus,
					OnFocusLost:    onFocusLost,
					OnRenderToMain: onRenderToMain,
				}),
				list:              viewModel,
				viewTrait:         NewViewTrait(view),
				getDisplayStrings: getDisplayStrings,
				c:                 c,
			},
		},
	}
}

func (self *LocalCommitsContext) GetSelectedItemId() string {
	item := self.GetSelected()
	if item == nil {
		return ""
	}

	return item.ID()
}

type LocalCommitsViewModel struct {
	*FilteredListViewModel[*models.Commit]

	// If this is true we limit the amount of commits we load, for the sake of keeping things fast.
	// If the user attempts to scroll past the end of the list, we will load more commits.
	limitCommits bool

	// If this is true we'll use git log --all when fetching the commits.
	showWholeGitGraph bool
}

func NewLocalCommitsViewModel(getModel func() []*models.Commit, c *types.HelperCommon, getNeedle func() string) *LocalCommitsViewModel {
	toString := func(commit *models.Commit) string {
		// TODO: include more stuff
		return commit.Name
	}

	self := &LocalCommitsViewModel{
		FilteredListViewModel: NewFilteredListViewModel(getModel, getNeedle, toString),
		limitCommits:          true,
		showWholeGitGraph:     c.UserConfig.Git.Log.ShowWholeGraph,
	}

	return self
}

func (self *LocalCommitsContext) CanRebase() bool {
	return true
}

func (self *LocalCommitsContext) GetSelectedRef() types.Ref {
	commit := self.GetSelected()
	if commit == nil {
		return nil
	}
	return commit
}

func (self *LocalCommitsViewModel) SetLimitCommits(value bool) {
	self.limitCommits = value
}

func (self *LocalCommitsViewModel) GetLimitCommits() bool {
	return self.limitCommits
}

func (self *LocalCommitsViewModel) SetShowWholeGitGraph(value bool) {
	self.showWholeGitGraph = value
}

func (self *LocalCommitsViewModel) GetShowWholeGitGraph() bool {
	return self.showWholeGitGraph
}

func (self *LocalCommitsViewModel) GetCommits() []*models.Commit {
	return self.getModel()
}

func cherryPickedCommitShaSet(state GuiContextState) *set.Set[string] {
	return models.ToShaSet(state.Modes().CherryPicking.CherryPickedCommits)
}
