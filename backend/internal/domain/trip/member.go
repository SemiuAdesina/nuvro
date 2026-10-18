package trip

type Role string

const (
	Owner  Role = "owner"
	Editor Role = "editor"
	Viewer Role = "viewer"
)

func (r Role) CanRead() bool          { return r == Owner || r == Editor || r == Viewer }
func (r Role) CanEdit() bool          { return r == Owner || r == Editor }
func (r Role) CanManageMembers() bool { return r == Owner }
