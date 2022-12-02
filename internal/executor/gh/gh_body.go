package gh

import (
	"bytes"
	"fmt"
	"text/template"
)

type (
	Body struct {
		Cluster      Cluster
		PodLogs      string
		PodName      string
		PodNamespace string
		PodDescribe  string
		Error        string
	}
	Cluster struct {
		Name    string
		Version string
	}
)

var bodyTpl = `
## Description

This issue refers to the problems connected with Pod ` + "`{{ .PodName }}` in namespace `{{ .PodNamespace }}` caused by the `{{ .Error }}` error." + `


<details>
  <summary><b>Pod logs</b></summary>

` + "```bash" + `
{{ .PodLogs }}
` + "```" + `
</details>

<details>
  <summary><b>Pod describe</b></summary>

` + "```yaml" + `
{{ .PodDescribe }}
` + "```" + `
</details>

### Cluster details

Name: **{{ .Cluster.Name}}**
Version: **{{ .Cluster.Version}}**

_Issue created via 'gh' Botkube plugin._
`

func RenderIssueBody(data *Body) (string, error) {
	tmpl, err := template.New("issue-body").
		Parse(bodyTpl)
	if err != nil {
		return "", fmt.Errorf("while creating template: %w", err)
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, data)
	if err != nil {
		return "", fmt.Errorf("while generating body: %w", err)
	}

	return body.String(), nil
}
