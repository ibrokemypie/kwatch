package source

type Bookmark struct {
	Backend  string
	Address  string
	Path     string
	Username string
	Password string
}

func (b Bookmark) Title() string {
	return b.Address + b.Path
}

func (b Bookmark) Description() string {
	return b.Backend
}

func (b Bookmark) FilterValue() string {
	return b.Description() + b.Title()
}
