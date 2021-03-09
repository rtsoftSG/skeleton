package generator

import (
	"bytes"
	"embed"
	"io"
	"os"
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
	if err := g.createDirectoryLayout(); err != nil {
		return err
	}

	rootDir := g.settings.ProjectRootDir

	if err := executeTplIntoFile(g.writeGoMod, path.Join(rootDir, "go.mod")); err != nil {
		return err
	}

	if err := executeTplIntoFile(g.writeMain, path.Join(rootDir, "cmd", g.settings.ProjectName, "main.go")); err != nil {
		return err
	}

	if err := executeTplIntoFile(g.writeLogger, path.Join(rootDir, "internal/infrastructure/logger/logger.go")); err != nil {
		return err
	}

	if err := executeTplIntoFile(g.writeConfig, path.Join(rootDir, "internal/config/config.go")); err != nil {
		return err
	}

	if err := executeTplIntoFile(g.writeConfigYml, path.Join(rootDir, "configs/config.yml")); err != nil {
		return err
	}

	if settings.UseJaeger {
		if err := executeTplIntoFile(g.writeTracer, path.Join(rootDir, "internal/infrastructure/tracer/jaeger.go")); err != nil {
			return err
		}
	}

	if err := executeTplIntoFile(g.writeApp, path.Join(rootDir, "internal/app.go")); err != nil {
		return err
	}

	return nil
}

func executeTplIntoFile(executor func(w io.Writer) error, filePath string) error {
	buff := &bytes.Buffer{}
	if err := executor(buff); err != nil {
		return err
	}
	return os.WriteFile(filePath, buff.Bytes(), 0644)
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
