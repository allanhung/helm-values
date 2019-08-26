[![Build Status](https://travis-ci.org/shihyuho/helm-values.svg?branch=master)](https://travis-ci.org/shihyuho/helm-values)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/shihyuho/helm-values/blob/master/LICENSE)
[![Build Status](https://github-basic-badges.herokuapp.com/release/shihyuho/helm-values.svg)](https://github.com/shihyuho/helm-values/releases)

# Helm Values Plugin

[Helm](https://github.com/helm/helm) doesn't support specify *value.yaml* when packaging chart archive. Therefore this plugin helps developers merge one or more YAML files of values for easily packaging different environments Helm Charts.

[![asciicast](https://asciinema.org/a/195778.png)](https://asciinema.org/a/195778)

## Install

Fetch the latest binary release of helm-values and install it:
 
```sh
$ helm plugin install https://github.com/allanhung/helm-values
```

## Usage
 
```sh
$ helm values [flags] CHART
```

### Flags

```sh
Flags:
      --backup-suffix string   suffix append to values.yaml if values.yaml already exist in output-dir (default ".bak")
  -h, --help                   help for helm
  -o, --output-dir string      writes the merged values to files in output-dir instead of stdout
  -f, --values valueFiles      specify values in a YAML file (can specify multiple) (default [])
  -t, --value-template string  specify value template in a YAML file
  -e, --env string             specify envirnment tag in template
  -s, --service string         specify service tag in template
```

## Example

The structure is like:

```sh
.
├── mychart
│   ├── .helmignore
│   ├── Chart.yaml
│   ├── charts
│   ├── templates
│   └── values.yaml
└── myenv
    ├── dev.yaml
    ├── sit.yaml
    └── uat.yaml
```
or

```sh
.
mychart
├── .helmignore
├── Chart.yaml
├── charts
├── templates
├── values.yaml
└── values_template.yaml
```

The script for package different environments chart archive:

```sh
# Merge sit values.yaml
$ helm values mychart --values myenv/sit.yaml --output-dir mychart

# Merge sit by template
$ helm values mychart -t values_template.yaml -e prod -s mysvc  --output-dir mychart

$ helm values mychart --values myenv/sit.yaml --output-dir mychart
# Package
$ helm package mychart

# Restore values.yaml
$ mv mychart/values.yaml.bak mychart/values.yaml
```
