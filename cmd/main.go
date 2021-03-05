package main

import (
	"fmt"
	"github.com/dixonwille/wlog/v3"
	"github.com/dixonwille/wmenu/v5"
	"github.com/rtsoftSG/skeleton/internal/generator"
	"log"
)

func main() {
	var generatorSettings generator.Settings

	generatorSettings.ProjectRootDir = "/home/kostya/go/src/skeleton/test"
	generatorSettings.ProjectName = "ptest"

	if err := runChooseConsulMenu(&generatorSettings); err != nil {
		log.Fatal(err)
	}
	if err := runChooseJaegerMenu(&generatorSettings); err != nil {
		log.Fatal(err)
	}
	if err := runChooseLoggerMenu(&generatorSettings); err != nil {
		log.Fatal(err)
	}
	if err := runChooseDBMenu(&generatorSettings); err != nil {
		log.Fatal(err)
	}

	fmt.Println(generatorSettings)

	if err := generator.Run(&generatorSettings); err != nil {
		log.Fatal(err)
	}

}

func runChooseConsulMenu(s *generator.Settings) error {
	defer fmt.Print("\n")

	consulMenu := wmenu.NewMenu("Use consul?")
	consulMenu.IsYesNo(wmenu.DefY)
	consulMenu.AddColor(wlog.None, wlog.BrightGreen, wlog.None, wlog.Red)

	consulMenu.Action(func(opts []wmenu.Opt) error {
		s.UseConsul = opts[0].Value.(string) == "yes"

		if s.UseConsul {
			m := wmenu.NewMenu("Sync config with consul?")
			m.IsYesNo(wmenu.DefY)
			m.AddColor(wlog.None, wlog.BrightGreen, wlog.None, wlog.Red)
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
	defer fmt.Print("\n")

	jaegerMenu := wmenu.NewMenu("Use jaeger tracer?")
	jaegerMenu.IsYesNo(wmenu.DefY)
	jaegerMenu.AddColor(wlog.None, wlog.BrightGreen, wlog.None, wlog.Red)

	jaegerMenu.Action(func(opts []wmenu.Opt) error {
		s.UseJaeger = opts[0].Value.(string) == "yes"
		return nil
	})

	return jaegerMenu.Run()
}

func runChooseLoggerMenu(s *generator.Settings) error {
	defer fmt.Print("\n")

	loggerMenu := wmenu.NewMenu("Choose logger")

	loggerMenu.LoopOnInvalid()
	loggerMenu.AddColor(wlog.None, wlog.BrightGreen, wlog.None, wlog.Red)

	loggerMenu.Action(func(opts []wmenu.Opt) error {
		s.Logger = opts[0].Value.(generator.LoggerChoice)
		return nil
	})

	loggerMenu.Option(string(generator.GoKit), generator.GoKit, true, nil)
	loggerMenu.Option(string(generator.Zap), generator.Zap, false, nil)
	return loggerMenu.Run()
}

func runChooseDBMenu(s *generator.Settings) error {
	defer fmt.Print("\n")

	dbMenu := wmenu.NewMenu("Choose database")

	dbMenu.LoopOnInvalid()
	dbMenu.AddColor(wlog.None, wlog.BrightGreen, wlog.None, wlog.Red)

	dbMenu.Action(func(opts []wmenu.Opt) error {
		s.Database = opts[0].Value.(generator.DBChoice)
		return nil
	})
	dbMenu.Option(string(generator.NoDb), generator.NoDb, true, nil)
	dbMenu.Option(string(generator.Clickhouse), generator.Clickhouse, false, nil)
	dbMenu.Option(string(generator.Postgresql), generator.Postgresql, false, nil)
	return dbMenu.Run()
}
