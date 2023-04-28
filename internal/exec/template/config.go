package template

import "strings"

type Source struct {
	Ref string `yaml:"ref"`
}

type (
	Config struct {
		Templates []Templates `yaml:"templates"`
	}
	Templates struct {
		Message InteractiveMessage `yaml:"message"`
		Command InteractiveCommand `yaml:"command"`
	}

	InteractiveCommand struct {
		Parser string `yaml:"parser"`
		Prefix string `yaml:"prefix"`
	}

	InteractiveMessage struct {
		Select  DropdownSelect    `yaml:"select"`
		Actions map[string]string `yaml:"actions"`
		Preview string            `yaml:"preview"`
	}

	DropdownSelect struct {
		Name    string `yaml:"name"`
		ItemKey string `yaml:"itemKey"`
		ItemIdx int    `json:"-"`
		Replace bool   `json:"-"`
	}
)

func (e Config) FindWithPrefix(cmd string) (Templates, bool) {
	for _, item := range e.Templates {
		if strings.HasPrefix(cmd, item.Command.Prefix) {
			return item, true
		}
	}

	return Templates{}, false
}
