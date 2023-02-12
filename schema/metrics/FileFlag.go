package metrics

import (
	"fmt"
	"log"
	"os"
)

type FileFlag struct {
	filename string

	Enabled bool
}

func NewFileFlag(filename string) *FileFlag {
	filename = fmt.Sprintf("%v/%v", os.TempDir(), filename)
	if _, err := os.Stat(filename); err == nil {
		os.Remove(filename)
	}
	return &FileFlag{
		filename: filename,
		Enabled:  false,
	}
}

func (fm *FileFlag) IsEnabled() bool {
	return fm.Enabled
}

func (fm *FileFlag) Enable() {
	if !fm.Enabled {
		fm.Enabled = true
		fm.OnValueChange(fm.Enabled)
	}
}

func (fm *FileFlag) Disable() {
	if fm.Enabled {
		fm.Enabled = false
		fm.OnValueChange(fm.Enabled)
	}
}

func (fm *FileFlag) OnValueChange(value bool) {
	if value {
		_, err := os.Create(fm.filename)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if _, err := os.Stat(fm.filename); err == nil {
			os.Remove(fm.filename)
		}
	}
}
