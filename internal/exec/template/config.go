package template

import "strings"

type (
	Config struct {
		Interactive []Interactive `yaml:"interactive"`
	}
	Interactive struct {
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

func (e Config) FindWithPrefix(cmd string) (Interactive, bool) {
	for _, item := range e.Interactive {
		if strings.HasPrefix(cmd, item.Command.Prefix) {
			return item, true
		}
	}

	return Interactive{}, false
}
