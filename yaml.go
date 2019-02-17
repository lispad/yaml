// Package yaml implements YAML support with extension for importing a sub-configuration tree for the Go language.
//
// Source code and other details for the project are available at GitHub:
//
//   https://github.com/lispad/yaml
//
package yaml

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
)

type (
	configImport struct {
		Resource     string `yaml:"resource"`
		IgnoreErrors bool   `yaml:"ignore_errors"`
		corrupted    bool
	}
	configImports struct {
		Imports []configImport `yaml:"imports"`
	}

	ReadFileFunc func(filename string) ([]byte, error)
)

var WrongDstTypeErr = errors.New("wrong type of dst argument: only pointer to struct is supported")

// ProcessFileWithImports processes config file and all it's imports tree
// Currently only pointer to struct is supported as dst argument
// Need to implement map merging(current version does full override) to support Map dst
func ProcessFileWithImports(configPath string, dst interface{}) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return WrongDstTypeErr
	}

	return processFile(configPath, dst, ioutil.ReadFile)
}

func processFile(configPath string, dst interface{}, reader ReadFileFunc) error {
	importList, err := getReverseOrderedImports(configPath, reader)
	if err != nil {
		return err
	}

	// process from the deepest imports to base file to allow override settings
	for i := len(importList) - 1; i >= 0; i-- {
		if importList[i].corrupted {
			continue
		}
		currentConfigRaw, readErr := reader(importList[i].Resource)
		if readErr != nil {
			if importList[i].IgnoreErrors {
				continue
			}
			return readErr
		}
		if yamlErr := yaml.Unmarshal(currentConfigRaw, dst); yamlErr != nil {
			if importList[i].IgnoreErrors {
				continue
			}
			return yamlErr
		}
	}

	return nil
}

func getReverseOrderedImports(configPath string, reader ReadFileFunc) ([]configImport, error) {
	var (
		configDir, _  = filepath.Split(configPath)
		importList    = []configImport{{Resource: configPath, IgnoreErrors: false}}
		currentConfig configImports
	)

	for i := 0; i < len(importList); i++ {
		currentConfigRaw, readErr := reader(importList[i].Resource)
		if readErr != nil {
			if importList[i].IgnoreErrors {
				importList[i].corrupted = true
				continue
			}
			return nil, readErr
		}
		if yamlErr := yaml.Unmarshal(currentConfigRaw, &currentConfig); yamlErr != nil {
			if importList[i].IgnoreErrors {
				importList[i].corrupted = true
				continue
			}
			return nil, yamlErr
		}
		for i := len(currentConfig.Imports) - 1; i >= 0; i-- {
			importFile := currentConfig.Imports[i]
			if !filepath.IsAbs(importFile.Resource) {
				importFile.Resource = configDir + importFile.Resource
			}
			importList = append(importList, importFile)
		}
		currentConfig.Imports = currentConfig.Imports[:0]
	}

	return importList, nil
}
