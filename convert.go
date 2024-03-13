package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

func main() {
	raw, err := os.ReadFile("/in/compose.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read compose file /in/compose.yaml")
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var model map[string]any
	err = yaml.Unmarshal(raw, &model)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to parse compose file /in/compose.yaml")
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = Convert(model, "/templates", "/out")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to apply template")
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func Convert(model map[string]any, templateDir string, out string) error {
	dir, err := os.ReadDir(templateDir)
	if err != nil {
		return fmt.Errorf("cannot access templates dir: %w", err)
	}
	for _, entry := range dir {
		f := filepath.Join(templateDir, entry.Name())
		if applyTemplate(model, f, out); err != nil {
			return err
		}
	}
	return nil
}

func applyTemplate(model map[string]any, file string, output string) error {
	tmpl, err := template.New(filepath.Base(file)).Funcs(helpers).ParseFiles(file)
	if err != nil {
		ExitError("cannot parse template "+file, err)
	}

	buff := bytes.Buffer{}
	err = tmpl.Execute(&buff, model)
	if err != nil {
		ExitError("cannot execute template "+file, err)
	}

	decoder := yaml.NewDecoder(&buff)
	for {
		var doc yaml.Node
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			ExitError("failed to parse generated yaml "+file, err)
		}

		out := bytes.Buffer{}
		encoder := yaml.NewEncoder(&out)
		err = encoder.Encode(&doc)
		if err != nil {
			ExitError("failed to parse generated yaml "+file, err)
		}
		fileOut := fileComment(&doc)
		if fileOut != "" {
			f := filepath.Join(output, fileOut)
			os.WriteFile(f, out.Bytes(), 0o700)
			fmt.Printf("Kubernetes resource \033[32;1m%s\033[0;m created\n", fileOut)
		} else {
			fmt.Println(out.String())
		}
	}
	return nil
}

func fileComment(node *yaml.Node) string {
	if node.HeadComment == "" {
		if len(node.Content) > 0 {
			return fileComment(node.Content[0])
		}
		return ""
	}
	for _, s := range strings.Split(node.HeadComment, "\n") {
		s := strings.TrimSpace(s)
		if strings.HasPrefix(s, "#! ") {
			return s[2:]
		}
	}
	return ""
}

var helpers = map[string]any{
	"required": func(attr string, a any) any {
		if a != nil {
			return a
		}
		ExitError("missing required attribute in compose model", errors.New(attr))
		return nil
	},
	"seconds": func(s any) float64 {
		duration, _ := time.ParseDuration(s.(string))
		return duration.Seconds()
	},
	"uppercase": func(s string) string {
		return strings.ToUpper(s)
	},
	"title": func(s string) string {
		return strings.Title(s)
	},
	"safe": func(s string) string {
		s = strings.ToLower(s)
		s = strings.Map(func(r rune) rune {
			if r < 'a' || r > 'z' {
				return '-'
			}
			return r
		}, s)
		for len(s) > 0 && s[0] == '-' {
			s = s[1:]
		}
		return s
	},
	"truncate": func(n int, s []any) []any {
		return s[n:]
	},
	"base64": func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	},
	"readfile": func(s string) string {
		file, err := os.ReadFile(s)
		if err != nil {
			ExitError("failed to read "+s, err)
		}
		return string(file)
	},
	"getenv": func(s string) string {
		return os.Getenv(s)
	},
	"dir": func(s string) string {
		return filepath.Dir(s)
	},
}

func ExitError(message string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", message, err)
	os.Exit(1)
}
