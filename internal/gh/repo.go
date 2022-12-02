package gh

// Repository represents a GitHub repository.
type Repository struct {
	owner    string
	name     string
	hostname string
}

// NewRepo instantiates a GitHub repository from owner and name arguments.
func NewRepo(owner, name string) Repository {
	return Repository{
		owner:    owner,
		name:     name,
		hostname: "github.com",
	}
}

func (r Repository) RepoOwner() string {
	return r.owner
}

func (r Repository) RepoName() string {
	return r.name
}

func (r Repository) RepoHost() string {
	return r.hostname
}
