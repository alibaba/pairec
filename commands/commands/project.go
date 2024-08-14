package commands

import (
	"os"
	"strings"

	"github.com/alibaba/pairec/v2/pairecmd/app/options"
	"github.com/alibaba/pairec/v2/pairecmd/log"
)

func Project(rootcfg *options.RootConfiguration, cfg *options.ProjectConfiguration) error {
	// create project dir
	if err := createDir(cfg.Name); err != nil {
		return err
	}

	// create src dir
	srcDir := cfg.Name + string(os.PathSeparator) + "src"
	if err := createDir(srcDir); err != nil {
		return err
	}

	// create src/controller dir
	if err := createDir(cfg.Name + string(os.PathSeparator) + "src" + string(os.PathSeparator) + "controller"); err != nil {
		return err
	}

	// create conf dir
	if err := createDir(cfg.Name + string(os.PathSeparator) + "conf"); err != nil {
		return err
	}

	// create docker dir
	if err := createDir(cfg.Name + string(os.PathSeparator) + "docker"); err != nil {
		return err
	}

	if err := createMakefile(cfg.Name); err != nil {
		return err
	}

	if err := createGomodfile(cfg.Name); err != nil {
		return err
	}

	if err := createMainfile(cfg.Name); err != nil {
		return err
	}

	if err := createFile(cfg.Name+string(os.PathSeparator)+"conf"+string(os.PathSeparator)+"config.json.production", confS); err != nil {
		return err
	}

	if err := createFile(cfg.Name+string(os.PathSeparator)+"docker"+string(os.PathSeparator)+"Dockerfile", dockerfileS); err != nil {
		return err
	}

	if err := createFile(cfg.Name+string(os.PathSeparator)+"src"+string(os.PathSeparator)+"controller"+string(os.PathSeparator)+"feed.go", controllerfileS); err != nil {
		return err
	}

	log.Info("create project " + cfg.Name + " finished.")

	return nil
}

func createDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func createMakefile(projectName string) error {

	file := projectName + string(os.PathSeparator) + "Makefile"
	bin := strings.ReplaceAll(projectName, "-", "_")

	content := strings.ReplaceAll(makefileS, "${BINNAME}", bin)

	return createFile(file, content)

}

func createGomodfile(projectName string) error {

	file := projectName + string(os.PathSeparator) + "go.mod"
	bin := strings.ReplaceAll(projectName, "-", "_")

	content := strings.ReplaceAll(gomodS, "${BINNAME}", bin)

	return createFile(file, content)
}

func createMainfile(projectName string) error {

	file := projectName + string(os.PathSeparator) + "src" + string(os.PathSeparator) + "main.go"
	bin := strings.ReplaceAll(projectName, "-", "_")

	content := strings.ReplaceAll(mainfileS, "${BINNAME}", bin)

	return createFile(file, content)
}

func createFile(file, content string) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return err
	}
	return nil
}
