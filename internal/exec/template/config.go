package template

import (
	"fmt"
	"strings"

	"github.com/kubeshop/botkube/pkg/api"
	"gopkg.in/yaml.v3"
)

type Source struct {
	Ref string `yaml:"ref"`
}

type (
	Config struct {
		Templates []Template `yaml:"templates"`
	}

	Templates []Template
	Template  struct {
		Type            string          `yaml:"type"`
		Trigger         Trigger         `yaml:"trigger"`
		ParseMessage    ParseMessage    `yaml:"-"`
		BuilderMessage  BuilderMessage  `yaml:"-"`
		WrapMessage     WrapMessage     `yaml:"-"`
		TutorialMessage TutorialMessage `yaml:"-"`
	}

	Trigger struct {
		Command string `yaml:"command"`
	}

	ParseMessage struct {
		Selects []Select          `yaml:"selects"`
		Actions map[string]string `yaml:"actions"`
		Preview string            `yaml:"preview"`
	}
	WrapMessage struct {
		Buttons api.Buttons `yaml:"buttons"`
	}
	BuilderMessage struct {
		Template string            `yaml:"template"`
		Selects  []Select          `yaml:"selects"`
		Actions  map[string]string `yaml:"actions"`
	}
	TutorialMessage struct {
		Buttons  api.Buttons `yaml:"buttons"`
		Header   string      `yaml:"header"`
		Paginate Paginate    `yaml:"paginate"`
	}
	Paginate struct {
		Page        int `yaml:"page"`
		CurrentPage int `yaml:"-"`
	}
	Select struct {
		Name   string `yaml:"name"`
		KeyTpl string `yaml:"keyTpl"`
	}
)

func (su *Template) UnmarshalYAML(node *yaml.Node) error {
	var data struct {
		Type    string  `yaml:"type"`
		Trigger Trigger `yaml:"trigger"`
	}
	err := node.Decode(&data)
	if err != nil {
		return err
	}

	switch {
	case strings.HasPrefix(data.Type, "builder"):
		var data struct {
			Message BuilderMessage `yaml:"message"`
		}
		err = node.Decode(&data)
		if err != nil {
			return err
		}
		su.BuilderMessage = data.Message
	case strings.HasPrefix(data.Type, "parser:"):
		var data struct {
			Message ParseMessage `yaml:"message"`
		}
		err = node.Decode(&data)
		if err != nil {
			return err
		}
		su.ParseMessage = data.Message
	case data.Type == "wrapper":
		var data struct {
			Message WrapMessage `yaml:"message"`
		}
		err = node.Decode(&data)
		if err != nil {
			return err
		}
		su.WrapMessage = data.Message
	case data.Type == "tutorial":
		var data struct {
			Message TutorialMessage `yaml:"message"`
		}
		err = node.Decode(&data)
		if err != nil {
			return err
		}
		su.TutorialMessage = data.Message
	}

	su.Type = data.Type
	su.Trigger = data.Trigger
	return nil
}

func (e Config) FindWithPrefix(cmd string) (Template, bool) {
	for idx := range e.Templates {
		item := e.Templates[idx]
		if item.Trigger.Command == "" {
			continue
		}

		fmt.Println(cmd)
		fmt.Println(item.Trigger.Command)
		if strings.HasPrefix(cmd, item.Trigger.Command) {
			return item, true
		}
	}

	return Template{}, false
}
