// Package yaml implements YAML support with extension for importing a sub-configuration tree for the Go language.
//
// Source code and other details for the project are available at GitHub:
//
//   https://github.com/lispad/yaml
//
package yaml

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestGetReverseOrderedImports(t *testing.T) {
	var fakeReaderNoFileError = errors.New("no such file")
	testCases := []struct {
		files           map[string][]byte
		testFile        string
		expectedImports []configImport
		expectedError   error
	}{
		// import order test cases
		{
			map[string][]byte{
				"config1.yml": []byte("no_imports: here"),
			},
			"config1.yml",
			[]configImport{
				{Resource: "config1.yml", corrupted: false, IgnoreErrors: false},
			},
			nil,
		},
		{
			map[string][]byte{
				"config1.yml": []byte("imports:\n - {resource: config2.yml}"),
				"config2.yml": []byte("no_imports: here"),
			},
			"config1.yml",
			[]configImport{
				{Resource: "config1.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "config2.yml", corrupted: false, IgnoreErrors: false},
			},
			nil,
		},
		{
			map[string][]byte{
				"config1.yml": []byte("imports:\n - {resource: config2.yml}\n - {resource: config3.yml}"),
				"config2.yml": []byte("no_imports: here"),
				"config3.yml": []byte("no_imports: here"),
			},
			"config1.yml",
			[]configImport{
				{Resource: "config1.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "config3.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "config2.yml", corrupted: false, IgnoreErrors: false},
			},
			nil,
		},
		{
			map[string][]byte{
				"config1.yml": []byte("imports:\n - {resource: config2.yml}\n - {resource: config3.yml}"),
				"config2.yml": []byte("imports:\n - {resource: config4.yml}"),
				"config3.yml": []byte("no_imports: here"),
				"config4.yml": []byte("no_imports: here"),
			},
			"config1.yml",
			[]configImport{
				{Resource: "config1.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "config3.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "config2.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "config4.yml", corrupted: false, IgnoreErrors: false},
			},
			nil,
		},
		// ignore_errors test cases
		{
			map[string][]byte{
				"config1.yml": []byte("imports:\n - {resource: config2.yml}"),
				"config2.yml": []byte("no_imports: here"),
			},
			"config1.yml",
			[]configImport{
				{Resource: "config1.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "config2.yml", corrupted: false, IgnoreErrors: false},
			},
			nil,
		},
		{
			map[string][]byte{
				"config1.yml": []byte("imports:\n - {resource: wrong_file.yml}"),
			},
			"config1.yml",
			nil,
			fakeReaderNoFileError,
		},
		{
			map[string][]byte{
				"config1.yml": []byte("imports:\n - {resource: wrong_file.yml, ignore_errors: true}"),
			},
			"config1.yml",
			[]configImport{
				{Resource: "config1.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "wrong_file.yml", corrupted: true, IgnoreErrors: true},
			},
			nil,
		},
		// abs path cases
		{
			map[string][]byte{
				"config/config1.yml":        []byte("imports:\n - {resource: config2.yml}\n - {resource: /abs/path/config3.yml}"),
				"config/config2.yml":        []byte("imports:\n - {resource: subdir/config4.yml}"),
				"/abs/path/config3.yml":     []byte("no_imports: here"),
				"config/subdir/config4.yml": []byte("no_imports: here"),
			},
			"config/config1.yml",
			[]configImport{
				{Resource: "config/config1.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "/abs/path/config3.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "config/config2.yml", corrupted: false, IgnoreErrors: false},
				{Resource: "config/subdir/config4.yml", corrupted: false, IgnoreErrors: false},
			},
			nil,
		},
		// corrupted cases
		{
			map[string][]byte{
				"config1.yml": []byte("not valid"),
			},
			"config1.yml",
			nil,
			&yaml.TypeError{Errors:[]string{"line 1: cannot unmarshal !!str `not valid` into yaml.configImports"}},
		},
	}

	for _, tc := range testCases {
		fakeReader := func(filename string) ([]byte, error) {
			if data, ok := tc.files[filename]; ok {
				return data, nil
			} else {
				return nil, fakeReaderNoFileError
			}
		}
		imports, err := getReverseOrderedImports(tc.testFile, fakeReader)
		assert.Equal(t, tc.expectedImports, imports)
		assert.Equal(t, tc.expectedError, err)

	}
}

func TestProcessFile(t *testing.T) {
	var fakeReaderNoFileError = errors.New("no such file")
	fakeReader := func(filename string) ([]byte, error) {
		switch filename {

		case "config1.yml":
			return []byte("imports:\n" +
				" - {resource: config2.yml}\n" +
				"a: config1, final value\n"), nil
		case "config2.yml":
			return []byte("imports:\n" +
				" - {resource: config3.yml}\n" +
				" - {resource: wrong_file.yaml, ignore_errors: true}\n" +
				"a: config2, will be overwritten again\n" +
				"b:\n" +
				" c: C value from config 2"), nil
		case "config3.yml":
			return []byte("" +
				"a: config3, will be overwritten twice\n" +
				"b:\n" +
				" c: will be overwritten once\n" +
				" d:\n" +
				"  e: will not be overwritten"), nil
		default:
			return nil, fakeReaderNoFileError
		}
	}

	type (
		nestedNestedStruct struct {
			E string
		}
		nestedStruct struct {
			C string
			D nestedNestedStruct
		}
		testStruct struct {
			A string
			B nestedStruct
		}
	)

	var ts, ts2, empty_ts testStruct
	expected := testStruct{
		A: "config1, final value",
		B: nestedStruct{
			C: "C value from config 2",
			D: nestedNestedStruct{
				E: "will not be overwritten",
			},
		},
	}

	err := processFile("config1.yml", &ts, fakeReader)
	assert.Nil(t, err)
	assert.Equal(t, expected, ts)

	err = processFile("wrong_file.yml", &ts2, fakeReader)
	assert.Equal(t, empty_ts, ts2)
	assert.Equal(t, fakeReaderNoFileError, err)
}

func TestProcessFileWithImports(t *testing.T) {
	unsupported := make(map[string]string)
	err := ProcessFileWithImports("any.yml", &unsupported)
	assert.Equal(t, WrongDstTypeErr, err, "wrong behaviour: expected to get WrongDstTypeErr when providing map")
}
