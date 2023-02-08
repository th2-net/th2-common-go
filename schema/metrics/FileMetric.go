package metrics

import (
	"fmt"
	"log"
	"os"
)

type FileMetric struct {
	filename string

	Enabled bool
}

func NewFileMetric(filename string) *FileMetric {
	filename = fmt.Sprintf("%v/%v", os.TempDir(), filename)
	if _, err := os.Stat(filename); err == nil {
		os.Remove(filename)
	}
	return &FileMetric{
		filename: filename,
		Enabled:  false,
	}
}

func (fm *FileMetric) IsEnabled() bool {
	return fm.Enabled
}

func (fm *FileMetric) Enable() {
	if !fm.Enabled {
		fm.Enabled = true
		fm.OnValueChange(fm.Enabled)
	}
}

func (fm *FileMetric) Disable() {
	if fm.Enabled {
		fm.Enabled = false
		fm.OnValueChange(fm.Enabled)
	}
}

func (fm *FileMetric) OnValueChange(value bool) {
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
