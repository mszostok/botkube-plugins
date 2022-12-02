package gh

import (
	"fmt"

	cliapi "github.com/cli/cli/v2/api"
	prShared "github.com/cli/cli/v2/pkg/cmd/pr/shared"
	gogh "github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
)

type Client struct {
	ghCli *cliapi.Client
}

func NewClient(token string) (*Client, error) {
	client, err := gogh.HTTPClient(&api.ClientOptions{
		AuthToken:   token,
		EnableCache: false,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		ghCli: cliapi.NewClientFromHTTP(client),
	}, nil
}

type CreateIssueInput struct {
	Title string
	Body  string
}

func (c *Client) CreateIssue(baseRepo Repository, in CreateIssueInput) (*cliapi.Issue, error) {
	tb := prShared.IssueMetadataState{
		Type:  prShared.IssueMetadata,
		Title: in.Title,
		Body:  in.Body,
	}
	params := map[string]interface{}{
		"title": tb.Title, // TODO?
		"body":  tb.Body,
	}

	err := prShared.AddMetadataToIssueParams(c.ghCli, baseRepo, params, &tb)
	if err != nil {
		return nil, fmt.Errorf("while adding metadata: %w", err)
	}

	ghRepo, err := cliapi.GitHubRepo(c.ghCli, baseRepo)
	if err != nil {
		return nil, fmt.Errorf("while getting repo information: %w", err)
	}

	newIssue, err := cliapi.IssueCreate(c.ghCli, ghRepo, params)
	if err != nil {
		return nil, fmt.Errorf("while creating issue: %w", err)
	}

	return newIssue, nil
}

func (c *Client) GetIssues(repo Repository) (cliapi.Issue, error) {
	q := `query IssueSearch($query: String!) {
  search(type: ISSUE, last:1, query: $query) {
    nodes {
      ...on Issue {
        title
        number
      }
    }
  }
}
`

	variables := map[string]interface{}{
		"query": fmt.Sprintf("Issue for in:title is:open is:issue repo:%s/%s", repo.RepoOwner(), repo.RepoName()),
	}

	type response struct {
		Search struct {
			IssueCount int
			Nodes      []cliapi.Issue
			PageInfo   struct {
				HasNextPage bool
				EndCursor   string
			}
		}
	}

	var resp response
	err := c.ghCli.GraphQL(repo.RepoHost(), q, variables, &resp)
	if err != nil {
		return cliapi.Issue{}, err
	}

	if len(resp.Search.Nodes) == 0 {
		return cliapi.Issue{}, nil
	}

	// in query, we specify `last: 1`, so we should get more than 1 result anyway.
	return resp.Search.Nodes[0], nil
}
