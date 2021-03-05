package generator

import (
	"embed"
	"os"
	"path"
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

	return nil
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

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/config"), 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	err = os.Mkdir(path.Join(g.settings.ProjectRootDir, "internal/transport"), 0755)
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
