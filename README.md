# YAML support for the Go language

Introduction
------------

This yaml package uses [https://gopkg.in/yaml.v2](https://gopkg.in/yaml.v2), 
and adds support for Symfony's yaml extension for importing a sub-configuration tree.    

Compatibility
-------------

This yaml package supports most of YAML 1.1 and 1.2, the same as https://gopkg.in/yaml.v2 does. 

This yaml package partially supports Symfony's yaml extension for importing configurations 
with following restrictions:
 * `parameters` section and `%parameter%` macros are not supported
 * full path to global configurations are not  yet supported

Installation and usage
----------------------

The import path for the package is *github.com/lispad/yaml*.

To install it, run:

    go get github.com/lispad/yaml

API documentation
-----------------

Original's package API documentation:

  * [https://gopkg.in/yaml.v2](https://gopkg.in/yaml.v2)

Symfony's configuration import documentation:
 * [https://symfony.com/doc/current/configuration/configuration_organization.html](https://symfony.com/doc/current/configuration/configuration_organization.html)

API stability
-------------

The package is provided without warranties or conditions of any kind.


Example
-------
There are 3 config files:
`configs/config1.yaml`
```yaml
imports:
  - {resource: config2.yaml}

a: config1, final value
b:
  d:
    f: 314
  g: 1
```

`configs/config2.yaml`
```yaml
imports:
  - {resource: subdir/config3.yaml}
  - {resource: wrong_file.yaml, ignore_errors: true}

a: config2, will be overriden again
b:
  c: C value from config 2
```

`configs/subdir/config3.yaml`
```yaml
a: config3, will be overriden twice
b:
  c: will be overriden once
  d:
    e: will not be overriden
    f: 1054
  g: 3
```


```Go
package main

import (
	"fmt"
	"log"

	"github.com/lispad/yaml"
)

type T struct {
	A string
	B struct {
		C string
		D struct {
			E string
			F int
		}
		RenamedG int `yaml:"g"`
	}
}

func main() {
	t := T{}

	err := yaml.ProcessFileWithImports("configs/config1.yaml", &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%+v\n\n", t)

	err = yaml.ProcessFileWithImports("wrong_file.yaml", &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
```

This example will generate the following output:

```
{A:config1, final value B:{C:C values from config 2 D:{E:will not be overriden F:314} RenamedG:1}}

error: open wrong_file.yaml: no such file or directory
```

