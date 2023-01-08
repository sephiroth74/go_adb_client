package types

import "strings"

type DumpsysParser struct {
	Lines []string
}

func (b DumpsysParser) FindSections() []string {
	var sections []string
	var lineWasEmpty = false

	for index, line := range b.Lines {
		trimmedLine := strings.TrimSpace(line)
		lineSize := len(trimmedLine)
		lineIsEmpty := lineSize == 0

		if index == 0 {
			sections = append(sections, trimmedLine)
			continue
		}

		if lineWasEmpty && !lineIsEmpty {
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				sections = append(sections, trimmedLine)
			}
		}

		lineWasEmpty = lineIsEmpty
	}
	return sections
}

func (b DumpsysParser) FindSection(title string) *DumpSysSection {
	var sections []string
	var lineWasEmpty = false
	var started = false

	for index, line := range b.Lines {
		trimmedLine := strings.TrimSpace(line)
		lineSize := len(trimmedLine)
		lineIsEmpty := lineSize == 0

		if index == 0 || (lineWasEmpty && !lineIsEmpty) {
			if !strings.HasPrefix(line, " ") {
				if strings.EqualFold(trimmedLine, title) {
					started = true
					continue
				} else if started {
					break
				}
			}
		}

		if started {
			sections = append(sections, trimmedLine)
		}

		lineWasEmpty = lineIsEmpty
	}

	if started {
		return &DumpSysSection{Title: title, Lines: sections}
	} else {
		return nil
	}

}

type DumpSysSection struct {
	Title string
	Lines []string
}
