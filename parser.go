package drushmakeparser

import (
	"bufio"
	"fmt"
	"github.com/go-ini/ini"
	"io/ioutil"
	"regexp"
	"strings"
)

type DrushMakeInfo struct {
	Packages []Package
}

type Package struct {
	Name     string
	Version  string
	Type     string
	Revision string
}

func (d *DrushMakeInfo) GetPackageByName(name string) Package {
	p := Package{}
	return p
}

func (d *DrushMakeInfo) GetPackageListByPrefix(prefix string) []Package {
	p := []Package{}
	return p
}

// Parse a makefile and populate the manifest.
func (d *DrushMakeInfo) Parse(filePath string) error {
	cfg, err := ini.Load(filePath)
	if err != nil {
		return err
	}

	// Core is parsed directly via ini read.
	core := Package{
		Name:    "drupal",
		Version: cfg.Section("").Key("core").Value(),
		Type:    "core",
	}

	d.Packages = append(d.Packages, core)

	// For all other packages we have to manually parse each line.
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	componentList := componentList(string(b))

	for _, name := range componentList {
		p := Package{}
		p.Parse(findBlock(name, string(b)))
		d.Packages = append(d.Packages, p)
	}
	return nil
}

// Parse package information out of a single block of text.
func (p *Package) Parse(block string) error {
	var patterns = map[string]string{
		"BASIC":    `^projects\[(\w+)\]\s?=\s?(\S+)$`,                         // projects[views] = 3.14
		"VERSION":  `^projects\[(\w+)\]\[version\]\s?=\s?(\S+)$`,              // projects[nodequeue][version] = 2.0-alpha1
		"BRANCH":   `^projects\[(\w+)\]\[download\]\[branch\]\s?=\s?(\S+)$`,   // projects[ns_core][download][branch] = 7.x-2.x
		"TYPE":     `^projects\[(\w+)\]\[download\]\[type\]\s?=\s?(\S+)$`,     // projects[ns_core][download][type] = git
		"REVISION": `^projects\[(\w+)\]\[download\]\[revision\]\s?=\s?(\S+)$`, // projects[draggableviews][download][revision] = 9677bc18b7255e13c33ac3cca48732b855c6817d
	}

	p.Name = keyMapper(block)

	// We assume that a single block will reference a single component, see the continue.
	scanner := bufio.NewScanner(strings.NewReader(block))
	for scanner.Scan() {
		line := scanner.Text()

		for rowType, expression := range patterns {
			re := regexp.MustCompile(expression)
			if isMatch := re.MatchString(line); isMatch {
				matches := re.FindStringSubmatch(line)

				// Sanity check for entries not for a single project. All regex captures the name as match[1].
				if matches[1] != p.Name {
					continue
				}

				// Populate the right components within the struct.
				switch rowType {
				case "BASIC", "VERSION", "BRANCH":
					p.Version = matches[2]
				case "TYPE":
					p.Type = matches[2]
				case "REVISION":
					p.Revision = matches[2]
				}
			}
		}
	}
	return nil
}

// Find a code block which contains a certain component key.
func findBlock(name string, rawBlock string) (componentBlock string) {
	scanner := bufio.NewScanner(strings.NewReader(rawBlock))
	componentBlock = ""

	for scanner.Scan() {
		line := scanner.Text()
		// Find a line which contains our keyword.
		if match := strings.Index(line, fmt.Sprintf("projects[%s]", name)); match > -1 {
			componentBlock += line + "\n"
		}
	}
	return componentBlock
}

// Helper function to extract the project name from a component block.
func keyMapper(key string) string {
	match, start, end := strings.Index(key, "projects["), strings.Index(key, "["), strings.Index(key, "]")
	if match == 0 && start > 0 && end > 0 {
		return key[start+1 : end]
	}
	return ""
}

// Build a list of deduped project names out of a raw projects[name] block.
func componentList(rawBlock string) (componentList []string) {
	components := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(rawBlock))

	for scanner.Scan() {
		key := keyMapper(scanner.Text())

		if key == "" || key == "drupal" { // Skip empty, or core lines.
			continue
		} else if _, no := components[key]; key != "" && !bool(no) {
			components[key] = true
		}
	}
	for name := range components {
		componentList = append(componentList, name)
	}
	return
}
