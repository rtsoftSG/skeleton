package main

import (
	"fmt"
	"github.com/dixonwille/wlog/v3"
	"github.com/dixonwille/wmenu/v5"
	"github.com/rtsoftSG/skeleton/internal/generator"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	var generatorSettings generator.Settings

	app := &cli.App{
		Name:                 "skeleton",
		Usage:                "A-PLATFORM microservice skeleton generator",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "generate skeleton code",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "directory",
						Aliases:  []string{"d"},
						Usage:    "`PATH` to directory where the new application will be created",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "application `NAME`",
						Required: true,
					},
					&cli.BoolFlag{
						Name:    "with-dependencies",
						Aliases: []string{"wd"},
						Usage:   "download service dependencies in vendor directory",
						Value:   true,
					},
				},
				Action: func(c *cli.Context) error {
					generatorSettings.ProjectRootDir = c.String("directory")
					if _, err := os.Stat(generatorSettings.ProjectRootDir); os.IsNotExist(err) {
						return fmt.Errorf("directory %s not exists", generatorSettings.ProjectRootDir)
					}
					generatorSettings.ProjectName = c.String("name")
					generatorSettings.WithDeps = c.Bool("with-dependencies")

					if err := runChooseConsulMenu(&generatorSettings); err != nil {
						return err
					}
					if err := runChooseJaegerMenu(&generatorSettings); err != nil {
						return err
					}
					if err := runChoosePrometheusMenu(&generatorSettings); err != nil {
						return err
					}
					if err := runChooseLoggerMenu(&generatorSettings); err != nil {
						return err
					}
					if err := runChooseDBMenu(&generatorSettings); err != nil {
						return err
					}
					if err := runChooseRouterMenu(&generatorSettings); err != nil {
						return err
					}

					return generator.Run(&generatorSettings)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func runChooseConsulMenu(s *generator.Settings) error {
	consulMenu := wmenu.NewMenu("Use consul?")
	consulMenu.IsYesNo(wmenu.DefY)
	consulMenu.AddColor(wlog.BrightGreen, wlog.BrightYellow, wlog.None, wlog.Red)

	consulMenu.Action(func(opts []wmenu.Opt) error {
		s.UseConsul = opts[0].Value.(string) == "yes"

		if s.UseConsul {
			m := wmenu.NewMenu("Sync config with consul?")
			m.IsYesNo(wmenu.DefY)
			consulMenu.AddColor(wlog.BrightGreen, wlog.BrightYellow, wlog.None, wlog.Red)
			m.Action(func(opts []wmenu.Opt) error {
				s.SyncConfigWithConsul = opts[0].Value.(string) == "yes"
				return nil
			})
			return m.Run()
		}

		return nil
	})

	return consulMenu.Run()
}

func runChooseJaegerMenu(s *generator.Settings) error {
	jaegerMenu := wmenu.NewMenu("Use jaeger tracer?")
	jaegerMenu.IsYesNo(wmenu.DefY)
	jaegerMenu.AddColor(wlog.BrightGreen, wlog.BrightYellow, wlog.None, wlog.Red)

	jaegerMenu.Action(func(opts []wmenu.Opt) error {
		s.UseJaeger = opts[0].Value.(string) == "yes"
		return nil
	})

	return jaegerMenu.Run()
}

func runChoosePrometheusMenu(s *generator.Settings) error {
	prometheusMenu := wmenu.NewMenu("Use prometheus?")
	prometheusMenu.IsYesNo(wmenu.DefY)
	prometheusMenu.AddColor(wlog.BrightGreen, wlog.BrightYellow, wlog.None, wlog.Red)

	prometheusMenu.Action(func(opts []wmenu.Opt) error {
		s.UsePrometheus = opts[0].Value.(string) == "yes"
		return nil
	})

	return prometheusMenu.Run()
}

func runChooseLoggerMenu(s *generator.Settings) error {
	loggerMenu := wmenu.NewMenu("Select logger")

	loggerMenu.LoopOnInvalid()
	loggerMenu.AddColor(wlog.BrightGreen, wlog.BrightYellow, wlog.None, wlog.Red)

	loggerMenu.Action(func(opts []wmenu.Opt) error {
		s.Logger = opts[0].Value.(generator.LoggerChoice)
		return nil
	})

	loggerMenu.Option(string(generator.GoKit), generator.GoKit, true, nil)
	loggerMenu.Option(string(generator.Zap), generator.Zap, false, nil)
	return loggerMenu.Run()
}

func runChooseDBMenu(s *generator.Settings) error {
	dbMenu := wmenu.NewMenu("Select database")

	dbMenu.LoopOnInvalid()
	dbMenu.AddColor(wlog.BrightGreen, wlog.BrightYellow, wlog.None, wlog.Red)

	dbMenu.Action(func(opts []wmenu.Opt) error {
		s.Database = opts[0].Value.(generator.DBChoice)
		return nil
	})
	dbMenu.Option(string(generator.NoDb), generator.NoDb, true, nil)
	dbMenu.Option(string(generator.Clickhouse), generator.Clickhouse, false, nil)
	dbMenu.Option(string(generator.Postgresql), generator.Postgresql, false, nil)
	return dbMenu.Run()
}

func runChooseRouterMenu(s *generator.Settings) error {
	routerMenu := wmenu.NewMenu("Select router")

	routerMenu.LoopOnInvalid()
	routerMenu.AddColor(wlog.BrightGreen, wlog.BrightYellow, wlog.None, wlog.Red)

	routerMenu.Action(func(opts []wmenu.Opt) error {
		s.Router = opts[0].Value.(generator.RouterChoice)
		return nil
	})
	routerMenu.Option(string(generator.GorillaMux), generator.GorillaMux, true, nil)
	routerMenu.Option(string(generator.GIN), generator.GIN, false, nil)
	return routerMenu.Run()
}
