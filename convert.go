package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

func main() {
	licenseAgreement := os.Getenv("LICENSE_AGREEMENT")
	agreement, err := strconv.ParseBool(licenseAgreement)
	if err != nil || !agreement {
		fmt.Fprintln(os.Stderr, "setup the LICENSE_AGREEMENT environment variable to accept the Docker Subscription Service Agreement")
		os.Exit(1)
	}
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
		if entry.Name() == "_index.tmpl" {
			continue
		}
		f := filepath.Join(templateDir, entry.Name())
		newOut := filepath.Join(out, entry.Name())
		if entry.IsDir() {
			err := os.MkdirAll(newOut, fs.ModePerm)
			if err != nil && !os.IsExist(err){
				return err
			}
			if err := Convert(model, f, newOut); err != nil {
				return err
			}
			continue
		}
		err := applyTemplate(model, f, out)
		if err != nil {
			return err
		}
	}
	index := filepath.Join(templateDir, "_index.tmpl")
	if _, err := os.Stat(index); err == nil {
		files, err := os.ReadDir(out)
		if err != nil {
			return err
		}
		err = applyTemplate(files, index, out)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyTemplate(model any, file string, output string) error {
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
		cleanOut := strings.ReplaceAll(out.String(), "⌦", "{{")
		cleanOut = strings.ReplaceAll(cleanOut, "⌫", "}}")

		fileOut := fileComment(&doc)
		if fileOut != "" {
			f := filepath.Join(output, fileOut)
			os.WriteFile(f, []byte(cleanOut), 0o700)
			fmt.Printf("Kubernetes resource \033[32;1m%s\033[0;m created\n", fileOut)
		} else {
			fmt.Println(cleanOut)
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
			return s[3:]
		}
	}
	return ""
}

var helpers = map[string]any{
	"helmValue": func(s string, args ...any) string {
		return fmt.Sprintf("⌦ %s ⌫", fmt.Sprintf(s, args...))
	},
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
		return safe(s)
	},
	"truncate": func(n int, s []any) []any {
		return s[n:]
	},
	"join": func(sep string, s []any) string {
		var ss []string
		for _, a := range s {
			ss = append(ss, a.(string))
		}
		return strings.Join(ss, sep)
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
	"indent": func(s string, indent int) string {
		indentation := strings.Repeat(" ", indent)
		lines := strings.Builder{}
		sc := bufio.NewScanner(strings.NewReader(s))
		for sc.Scan() {
			lines.WriteString(indentation)
			lines.WriteString(sc.Text())
			lines.WriteString("\n")
		}
		return lines.String()
	},
	"map": func(s string, rules ...string) string {
		for _, rule := range rules {
			before, after, _ := strings.Cut(rule, "->")
			if s == strings.TrimSpace(before) {
				return strings.TrimSpace(after)
			}
		}
		return s
	},
	"portName": func(service string, port any) string {
		var portAsString string
		switch port.(type) {
		case string:
			portAsString = port.(string)
			break
		case int:
			portAsString = strconv.Itoa(port.(int))
			break
		}
		shrinkTo := 15 - (len(portAsString) + 1)
		if len(service) < shrinkTo {
			shrinkTo = len(service)
		}
		return safe(fmt.Sprintf("%s-%s", service[0:shrinkTo], portAsString))
	},
}

func safe(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if ('a' <= r && r <= 'z') ||
			('A' <= r && r <= 'Z') ||
			('0' <= r && r <= '9') {
			return r
		}
		return '-'
	}, s)
	for len(s) > 0 && s[0] == '-' {
		s = s[1:]
	}
	return s
}

func ExitError(message string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", message, err)
	os.Exit(1)
}
