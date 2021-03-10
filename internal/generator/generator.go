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
		return err
	}

	rootDir := g.settings.ProjectRootDir

	log.Print("create go.mod file ...")
	if err := execTpl(g.writeGoMod, path.Join(rootDir, "go.mod")); err != nil {
		return err
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

	log.Print("create app.go ...")
	if err := execTplAndFormat(g.writeApp, path.Join(rootDir, "internal/app.go")); err != nil {
		return err
	}

	log.Print("create endpoint package ...")
	if err := execTplAndFormat(g.writeEndpoints, path.Join(rootDir, "internal/endpoint/endpoints.go")); err != nil {
		return err
	}
	if err := execTplAndFormat(g.writeEndpointsMiddlewares, path.Join(rootDir, "internal/endpoint/middleware.go")); err != nil {
		return err
	}
	if err := execTplAndFormat(g.writeEndpointsResponseRequest, path.Join(rootDir, "internal/endpoint/request.go")); err != nil {
		return err
	}

	log.Print("create transport/http package ...")
	if err := execTplAndFormat(g.writeHttpServer, path.Join(rootDir, "internal/transport/http/server.go")); err != nil {
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

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/config"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/endpoint"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
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

func (g *generator) writeGoMod(w io.Writer) error {
	tpl, err := template.New("gomod").ParseFS(templates, "assets/gomod")

	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{"module": g.settings.ProjectName})
}

func (g *generator) writeMain(w io.Writer) error {
	tpl, err := template.New("main").ParseFS(templates, "assets/main")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"module":           g.settings.ProjectName,
		"use_gokit_logger": g.settings.Logger == GoKit,
		"use_zap_logger":   g.settings.Logger == Zap,
	})
}

func (g *generator) writeLogger(w io.Writer) error {
	tpl, err := template.New("logger").ParseFS(templates, "assets/logger")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"use_gokit_logger": g.settings.Logger == GoKit,
		"use_zap_logger":   g.settings.Logger == Zap,
	})
}

func (g *generator) writeConfig(w io.Writer) error {
	tpl, err := template.New("config").ParseFS(templates, "assets/config")
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
	tpl, err := template.New("config_yml").ParseFS(templates, "assets/config_yml")
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
	tpl, err := template.New("tracer").ParseFS(templates, "assets/tracer")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{})
}

func (g *generator) writeApp(w io.Writer) error {
	tpl, err := template.New("app").ParseFS(templates, "assets/app")
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
	})
}

func (g *generator) writeEndpoints(w io.Writer) error {
	tpl, err := template.New("endpoint").ParseFS(templates, "assets/endpoint")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{
		"use_jaeger": g.settings.UseJaeger,
	})
}

func (g *generator) writeEndpointsMiddlewares(w io.Writer) error {
	tpl, err := template.New("endpoint_middleware").ParseFS(templates, "assets/endpoint_middleware")
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

func (g *generator) writeEndpointsResponseRequest(w io.Writer) error {
	tpl, err := template.New("endpoint_req_resp").ParseFS(templates, "assets/endpoint_req_resp")
	if err != nil {
		return err
	}

	return tpl.Execute(w, map[string]interface{}{})
}

func (g *generator) writeHttpServer(w io.Writer) error {
	tpl, err := template.New("http_server").ParseFS(templates, "assets/http_server")
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
		"use_gorilla":      g.settings.Router == GorillaMux,
		"use_gin":          g.settings.Router == GIN,
	})
}
