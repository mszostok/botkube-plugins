package exec

import (
	"strings"
)

type (
	Config struct {
		Interactive InteractiveSources `yaml:"interactive"`
	}
	InteractiveSources struct {
		Templates []Template `yaml:"templates"`
	}

	Template struct {
		Ref string `yaml:"ref"`
	}

	Interactive struct {
		Interactive []InteractiveItem `yaml:"interactive"`
	}
	InteractiveItem struct {
		Message Message            `yaml:"message"`
		Command InteractiveCommand `yaml:"command"`
	}

	InteractiveCommand struct {
		Parser string `yaml:"parser"`
		Prefix string `yaml:"prefix"`
	}

	Message struct {
		Select  Select            `yaml:"select"`
		Actions map[string]string `yaml:"actions"`
		Preview string            `yaml:"preview"`
	}

	Select struct {
		Name    string `yaml:"name"`
		ItemKey string `yaml:"itemKey"`
		ItemIdx int    `json:"-"`
		Replace bool   `json:"-"`
	}
)

func (e Interactive) FindWithPrefix(cmd string) (InteractiveItem, bool) {
	for _, item := range e.Interactive {
		if strings.HasPrefix(cmd, item.Command.Prefix) {
			return item, true
		}
	}

	return InteractiveItem{}, false
}
