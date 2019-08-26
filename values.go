package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	defaultDirectoryPermission = 0755
	defaultValuesFilename			= "values.yaml"
)

type valuesCmd struct {
	chartPath    string
	values       valueFiles
	valTemplate  string
	sections     sections
	outputDir    string
	backupSuffix string
}

func (cmd *valuesCmd) run() error {
	var vv []byte
	var err error
	if cmd.valTemplate == "" {
		cmd.values.Insert(path.Join(cmd.chartPath, defaultValuesFilename), 0)
		vv, err = vals(cmd.values)
		if err != nil {
			return err
		}
	} else {
		vv, err = valsections(path.Join(cmd.chartPath, cmd.valTemplate), cmd.sections)
		if err != nil {
			return err
		}
	}
	if cmd.outputDir != "" {
		return writeToFile(cmd.outputDir, defaultValuesFilename, cmd.backupSuffix, vv)
	} else {
		fmt.Println(string(vv))
	}
	return nil
}

// write the <data> to <output-dir>/<name>, backup <name> to <name><backup-suffix> first if exist
func writeToFile(outputDir string, name string, backupSuffux string, data []byte) error {
	outfileName := strings.Join([]string{outputDir, name}, string(filepath.Separator))

	err := ensureDirectoryForFile(outfileName)
	if err != nil {
		return err
	}

	err = ensureFileBackedup(outfileName, backupSuffux)
	if err != nil {
		return err
	}

	f, err := os.Create(outfileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	fmt.Printf("wrote %s\n", outfileName)
	return nil
}

// check if the file backed up
func ensureFileBackedup(file string, backupSuffix string) error {
	if _, err := os.Stat(file); err == nil {
		backupPath := path.Join(path.Dir(file), filepath.Base(file)+backupSuffix)
		ensureFileBackedup(backupPath, backupSuffix)
		copy(file, backupPath)
		fmt.Printf("backed up %s to %s\n", file, backupPath)
	}
	return nil
}

func copy(src, dest string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	return nil
}

// check if the directory exists to create file. creates if don't exists
func ensureDirectoryForFile(file string) error {
	baseDir := path.Dir(file)
	_, err := os.Stat(baseDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(baseDir, defaultDirectoryPermission)
}

func getSection(template yaml.MapSlice, tag []string) (yaml.MapSlice, error) {
	if len(tag) > 1 {
		for _, item := range template {
			if item.Key == tag[0] {
				return getSection(item.Value.(yaml.MapSlice), tag[1: len(tag)])
			}
		}
	} else {
		for _, item := range template {
			if item.Key == tag[0] {
				return item.Value.(yaml.MapSlice), nil
			}
    }
	}
	return yaml.MapSlice{}, nil
}

// vals merges values from template sections specified via -s/--sections
func valsections(templateFile string, sections sections) ([]byte, error) {
	var bytes []byte
	var err error
	template := yaml.MapSlice{}
	base := yaml.MapSlice{}

	if strings.TrimSpace(templateFile) == "-" {
		bytes, err = ioutil.ReadAll(os.Stdin)
	} else {
		bytes, err = ioutil.ReadFile(templateFile)
	}

	if err != nil {
		return []byte{}, err
	}

	if err := yaml.Unmarshal(bytes, &template); err != nil {
		return []byte{}, fmt.Errorf("failed to parse %s: %s", templateFile, err)
	}

	// User specified a sections via -f/--sections
	for _, section := range sections {
		// currentMap := map[string]interface{}{}
		currentMap, err := getSection(template, strings.Split(section,"."))

		if err != nil {
			return []byte{}, err
		}

		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	return yaml.Marshal(base)
}

// vals merges values from files specified via -f/--values
func vals(valueFiles valueFiles) ([]byte, error) {
	base := yaml.MapSlice{}

	// User specified a values files via -f/--values
	for _, filePath := range valueFiles {
		// currentMap := map[string]interface{}{}
		currentMap := yaml.MapSlice{}

		var bytes []byte
		var err error
		if strings.TrimSpace(filePath) == "-" {
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			bytes, err = ioutil.ReadFile(filePath)
		}

		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	return yaml.Marshal(base)
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest yaml.MapSlice, src yaml.MapSlice) yaml.MapSlice {
	for _, item := range src {
		// If the key doesn't exist already, then just set the key to that value
		if exists := containsKey(dest, item.Key); !exists {
			dest = setValue(dest, item.Key, item.Value)
			continue
		}
		nextMap, ok := item.Value.(yaml.MapSlice)
		// If it isn't another map, overwrite the value
		if !ok {
			dest = setValue(dest, item.Key, item.Value)
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := getValue(dest, item.Key)
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest = setValue(dest, item.Key, item.Value)
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		merged := mergeValues(destMap, nextMap)
		dest = setValue(dest, item.Key, merged)
	}
	return dest
}

func containsKey(slice yaml.MapSlice, key interface{}) bool {
	for _, item := range slice {
		if item.Key == key {
			return true
		}
	}
	return false
}

func setValue(slice yaml.MapSlice, key, newValue interface{}) yaml.MapSlice {
	for i := 0; i < len(slice); i++ {
		if slice[i].Key == key { // if key exist in slice, replace it
			slice[i].Value = newValue
			return slice
		}
	}
	// If we got to this point, it is a new key in slice, so just add at the end of slice
	return append(slice, yaml.MapItem{Key: key, Value: newValue})
}

func getValue(slice yaml.MapSlice, key interface{}) (value yaml.MapSlice, ok bool) {
	for _, item := range slice {
		if item.Key == key {
			value, ok = item.Value.(yaml.MapSlice)
			return
		}
	}
	return
}

type valueFiles []string

func (v *valueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *valueFiles) Type() string {
	return "valueFiles"
}

func (v *valueFiles) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}
	return nil
}

type sections	 []string

func (v *sections) String() string {
	return fmt.Sprint(*v)
}

func (v *sections) Type() string {
	return "sections"
}

func (v *sections) Set(value string) error {
	for _, section := range strings.Split(value, ",") {
		*v = append(*v, section)
	}
	return nil
}

// insert value to the index of valueFiles, append at last if index < 0
func (v *valueFiles) Insert(value string, index int) error {
	if index < 0 {
		return v.Set(value)
	}
	for _, filePath := range strings.Split(value, ",") {
		vv := append((*v)[:index], append(valueFiles{filePath}, (*v)[index:]...)...)
		*v = vv
	}
	return nil
}
