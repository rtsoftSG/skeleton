package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

// content holds our static web server content.
//go:embed assets/*
var templates embed.FS

type generator struct {
	settings *Settings
}

func Run(settings *Settings) error {
	g := generator{settings: settings}

	log.Print("create directories ...")
	if err := g.createDirectoryLayout(); err != nil {
		return fmt.Errorf("create directory structure: %w", err)
	}

	rootDir := g.settings.ProjectRootDir

	log.Print("create go.mod file ...")
	if err := execTpl(g.writeGoMod, path.Join(rootDir, "go.mod")); err != nil {
		return err
	}

	log.Print("create .gitignore file ...")
	if err := execTpl(g.writeGitignore, path.Join(rootDir, ".gitignore")); err != nil {
		return err
	}

	log.Print("create Dockerfile file ...")
	if err := execTpl(g.writeDockerfile, path.Join(rootDir, "Dockerfile")); err != nil {
		return err
	}

	log.Print("create Makefile file ...")
	if err := execTpl(g.writeMakefile, path.Join(rootDir, "Makefile")); err != nil {
		return err
	}

	log.Print("create README.md ...")
	if err := execTpl(g.writeReadme, path.Join(rootDir, "README.md")); err != nil {
		return err
	}

	log.Print("create golangci-lint yml config file ...")
	{
		if err := execTpl(g.writeGOlangCILint, path.Join(rootDir, ".golangci.yml")); err != nil {
			return err
		}

		if err := execTpl(g.writeGOlangCILintErrCheckExcludes, path.Join(rootDir, ".errcheck_excludes.txt")); err != nil {
			return err
		}
	}

	log.Print("create main.go ...")
	if err := execTplAndFormat(g.writeMain, path.Join(rootDir, "cmd", g.settings.ProjectName, "main.go")); err != nil {
		return err
	}

	log.Print("create config package ...")
	if err := execTplAndFormat(g.writeConfig, path.Join(rootDir, "internal/config/config.go")); err != nil {
		return err
	}

	log.Print("create config.yml example ...")
	if err := execTpl(g.writeConfigYml, path.Join(rootDir, "configs/config.yml")); err != nil {
		return err
	}

	log.Print("create infrastructure/logger package ...")
	if err := execTplAndFormat(g.writeLogger, path.Join(rootDir, "internal/infrastructure/logger/logger.go")); err != nil {
		return err
	}

	if settings.UseJaeger {
		log.Print("create infrastructure/tracer package ...")
		if err := execTplAndFormat(g.writeTracer, path.Join(rootDir, "internal/infrastructure/tracer/jaeger.go")); err != nil {
			return err
		}
	}

	if settings.UseConsul {
		log.Print("create infrastructure/consul package ...")
		if err := execTplAndFormat(g.writeConsul, path.Join(rootDir, "internal/infrastructure/consul/consul.go")); err != nil {
			return err
		}
	}

	log.Print("create app.go ...")
	if err := execTplAndFormat(g.writeApp, path.Join(rootDir, "internal/app.go")); err != nil {
		return err
	}

	if g.settings.Router == GorillaMux {
		log.Print("create endpoint package ...")
		if err := execTplAndFormat(g.writeEndpoints, path.Join(rootDir, "internal/endpoint/endpoints.go")); err != nil {
			return err
		}
		if err := execTplAndFormat(g.writeEndpointsMiddlewares, path.Join(rootDir, "internal/endpoint/middleware.go")); err != nil {
			return err
		}
	}

	log.Print("create transport/http package ...")
	switch settings.Router {
	case GorillaMux:
		if err := execTplAndFormat(g.writeGoKitHttpServer, path.Join(rootDir, "internal/transport/http/server.go")); err != nil {
			return err
		}
	case GIN:
		if err := execTplAndFormat(g.writeGinHttpServer, path.Join(rootDir, "internal/transport/http/server.go")); err != nil {
			return err
		}
	}

	log.Print("create test package ...")
	if err := execTplAndFormat(g.writeTest, path.Join(rootDir, "test/app_test.go")); err != nil {
		return err
	}

	if settings.WithDeps {
		log.Print("download dependencies ...")
		cmd := exec.Command("go", "mod", "vendor")
		cmd.Dir = rootDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("go mod vendor command: %w", err)
		}
	}

	log.Print("DONE!")

	return nil
}

func execTpl(executor func(w io.Writer) error, filePath string) error {
	buff := &bytes.Buffer{}
	if err := executor(buff); err != nil {
		return fmt.Errorf("on exec template: %w, file: %s", err, filePath)
	}

	return os.WriteFile(filePath, buff.Bytes(), 0644)
}

func execTplAndFormat(executor func(w io.Writer) error, filePath string) error {
	buff := &bytes.Buffer{}
	if err := executor(buff); err != nil {
		return fmt.Errorf("on exec template: %w, file: %s", err, filePath)
	}

	source, err := format.Source(buff.Bytes())
	if err != nil {
		return fmt.Errorf("on format sources: %w, file: %s", err, filePath)
	}

	return os.WriteFile(filePath, source, 0644)
}

func (g *generator) createDirectoryLayout() error {
	err := os.Mkdir(path.Join(g.settings.ProjectRootDir, "cmd/"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "cmd/", g.settings.ProjectName), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "test/"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/infrastructure"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/infrastructure/logger"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	if g.settings.UseJaeger {
		err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/infrastructure/tracer"), 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}

	if g.settings.UseConsul {
		err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/infrastructure/consul"), 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/config"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	if g.settings.Router == GorillaMux {
		err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/endpoint"), 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/transport"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/transport/http"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "configs"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "vendor"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}

func (g *generator) createTemplate(fileName string) (*template.Template, error) {
	return template.New(fileName).Funcs(template.FuncMap{
		"log": makeLogFunc(g.settings.Logger),
	}).ParseFS(templates, "assets/"+fileName)
}

func (g *generator) writeGoMod(w io.Writer) error {
	tpl, err := g.createTemplate("gomod")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{"module": g.settings.ProjectName})
}

func (g *generator) writeGitignore(w io.Writer) error {
	tpl, err := g.createTemplate("gitignore")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{})
}

func (g *generator) writeDockerfile(w io.Writer) error {
	tpl, err := g.createTemplate("dockerfile")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{"module": g.settings.ProjectName})
}

func (g *generator) writeMakefile(w io.Writer) error {
	tpl, err := g.createTemplate("makefile")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{"module": g.settings.ProjectName})
}

func (g *generator) writeGOlangCILint(w io.Writer) error {
	tpl, err := g.createTemplate("golangci_cfg")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{})
}

func (g *generator) writeGOlangCILintErrCheckExcludes(w io.Writer) error {
	tpl, err := g.createTemplate("errcheck_excludes")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{})
}

func (g *generator) writeReadme(w io.Writer) error {
	tpl, err := g.createTemplate("readme")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module":          strings.ToUpper(g.settings.ProjectName),
		"use_clickhouse":  g.settings.Database == Clickhouse,
		"use_postgresql":  g.settings.Database == Postgresql,
		"use_gorilla_mux": g.settings.Router == GorillaMux,
		"use_gin":         g.settings.Router == GIN,
		"use_jaeger":      g.settings.UseJaeger,
		"use_consul":      g.settings.UseConsul,
		"use_prometheus":  g.settings.UsePrometheus,
	})
}

func (g *generator) writeMain(w io.Writer) error {
	tpl, err := g.createTemplate("main")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module":           g.settings.ProjectName,
		"use_gokit_logger": g.settings.Logger == GoKit,
		"use_zap_logger":   g.settings.Logger == Zap,
		"use_jaeger":       g.settings.UseJaeger,
		"use_consul":       g.settings.UseConsul,
	})
}

func (g *generator) writeLogger(w io.Writer) error {
	tpl, err := g.createTemplate("logger")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"use_gokit_logger": g.settings.Logger == GoKit,
		"use_zap_logger":   g.settings.Logger == Zap,
	})
}

func (g *generator) writeConfig(w io.Writer) error {
	tpl, err := g.createTemplate("config")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module":                       g.settings.ProjectName,
		"use_clickhouse":               g.settings.Database == Clickhouse,
		"use_postgresql":               g.settings.Database == Postgresql,
		"use_jaeger":                   g.settings.UseJaeger,
		"use_consul":                   g.settings.UseConsul,
		"use_consul_for_configuration": g.settings.SyncConfigWithConsul,
	})
}

func (g *generator) writeConfigYml(w io.Writer) error {
	tpl, err := g.createTemplate("config_yml")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module":         g.settings.ProjectName,
		"use_clickhouse": g.settings.Database == Clickhouse,
		"use_postgresql": g.settings.Database == Postgresql,
		"use_jaeger":     g.settings.UseJaeger,
		"use_consul":     g.settings.UseConsul,
	})
}

func (g *generator) writeTracer(w io.Writer) error {
	tpl, err := g.createTemplate("tracer")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{})
}

func (g *generator) writeConsul(w io.Writer) error {
	tpl, err := g.createTemplate("consul")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{})
}

func (g *generator) writeApp(w io.Writer) error {
	tpl, err := g.createTemplate("app")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module":           g.settings.ProjectName,
		"use_clickhouse":   g.settings.Database == Clickhouse,
		"use_postgresql":   g.settings.Database == Postgresql,
		"use_gokit_logger": g.settings.Logger == GoKit,
		"use_zap_logger":   g.settings.Logger == Zap,
		"use_gorilla_mux":  g.settings.Router == GorillaMux,
		"use_gin":          g.settings.Router == GIN,
	})
}

func (g *generator) writeEndpoints(w io.Writer) error {
	tpl, err := g.createTemplate("endpoint_gokit")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"use_jaeger": g.settings.UseJaeger,
	})
}

func (g *generator) writeEndpointsMiddlewares(w io.Writer) error {
	tpl, err := g.createTemplate("endpoint_middleware")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module":           g.settings.ProjectName,
		"use_jaeger":       g.settings.UseJaeger,
		"use_consul":       g.settings.UseConsul,
		"use_gokit_logger": g.settings.Logger == GoKit,
		"use_zap_logger":   g.settings.Logger == Zap,
	})
}

func (g *generator) writeGoKitHttpServer(w io.Writer) error {
	tpl, err := g.createTemplate("http_server_gokit")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module":           g.settings.ProjectName,
		"use_clickhouse":   g.settings.Database == Clickhouse,
		"use_postgresql":   g.settings.Database == Postgresql,
		"use_jaeger":       g.settings.UseJaeger,
		"use_consul":       g.settings.UseConsul,
		"use_gokit_logger": g.settings.Logger == GoKit,
		"use_zap_logger":   g.settings.Logger == Zap,
		"use_prometheus":   g.settings.UsePrometheus,
	})
}

func (g *generator) writeGinHttpServer(w io.Writer) error {
	tpl, err := g.createTemplate("http_server_gin")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module":           g.settings.ProjectName,
		"use_clickhouse":   g.settings.Database == Clickhouse,
		"use_postgresql":   g.settings.Database == Postgresql,
		"use_jaeger":       g.settings.UseJaeger,
		"use_consul":       g.settings.UseConsul,
		"use_gokit_logger": g.settings.Logger == GoKit,
		"use_zap_logger":   g.settings.Logger == Zap,
		"use_prometheus":   g.settings.UsePrometheus,
	})
}

func (g *generator) writeTest(w io.Writer) error {
	tpl, err := g.createTemplate("app_test")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module": g.settings.ProjectName,
	})
}

func makeLogFunc(logger LoggerChoice) func(logger, lvl string, msg string) string {
	switch logger {
	case GoKit:
		return func(logger, lvl, msg string) string {
			return "level." + strings.Title(strings.ToLower(lvl)) + "(" + logger + ").Log(\"msg\", \"" + msg + "\")"
		}
	case Zap:
		return func(logger, lvl, msg string) string {
			return logger + "." + strings.Title(strings.ToLower(lvl)) + "(\"" + msg + "\")"
		}
	default:
		panic("unknown logger " + logger)
	}
}
