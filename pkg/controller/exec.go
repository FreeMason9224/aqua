package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/aqua/pkg/log"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

var (
	errCommandIsRequired = errors.New("command is required")
	errCommandIsNotFound = errors.New("command is not found")
)

func (ctrl *Controller) Exec(ctx context.Context, param *Param, args []string) error {
	if len(args) == 0 {
		return errCommandIsRequired
	}

	exeName := filepath.Base(args[0])
	fields := logrus.Fields{
		"exe_name": exeName,
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get the current directory: %w", logerr.WithFields(err, fields))
	}

	binDir := filepath.Join(ctrl.RootDir, "bin")

	if cfgFilePath := ctrl.getConfigFilePath(wd, param.ConfigFilePath); cfgFilePath != "" {
		pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
		if err != nil {
			return err
		}
		if pkg != nil {
			return ctrl.installAndExec(ctx, pkgInfo, pkg, file, binDir, args)
		}
	}
	cfgFilePath := ctrl.ConfigFinder.FindGlobal(ctrl.RootDir)
	if _, err := os.Stat(cfgFilePath); err != nil {
		return ctrl.findAndExecExtCommand(ctx, exeName, args[1:])
	}

	pkg, pkgInfo, file, err := ctrl.findExecFile(ctx, cfgFilePath, exeName)
	if err != nil {
		return err
	}
	if pkg == nil {
		return ctrl.findAndExecExtCommand(ctx, exeName, args[1:])
	}
	return ctrl.installAndExec(ctx, pkgInfo, pkg, file, binDir, args)
}

func (ctrl *Controller) findAndExecExtCommand(ctx context.Context, exeName string, args []string) error {
	exePath := lookPath(exeName)
	if exePath == "" {
		return logerr.WithFields(errCommandIsNotFound, logrus.Fields{ //nolint:wrapcheck
			"exe_name": exeName,
		})
	}
	return ctrl.execCommand(ctx, exePath, args)
}

func (ctrl *Controller) installAndExec(ctx context.Context, pkgInfo PackageInfo, pkg *Package, file *File, binDir string, args []string) error {
	fileSrc, err := pkgInfo.GetFileSrc(pkg, file)
	if err != nil {
		return fmt.Errorf("get file_src: %w", err)
	}

	if err := ctrl.installPackage(ctx, pkgInfo, pkg, binDir, false, false); err != nil {
		return err
	}

	return ctrl.exec(ctx, pkg, pkgInfo, fileSrc, args[1:])
}

func (ctrl *Controller) findExecFileFromPkg(registries map[string]*RegistryContent, exeName string, pkg *Package) (*Package, PackageInfo, *File) {
	registry, ok := registries[pkg.Registry]
	if !ok {
		log.New().Warnf("registry isn't found %s", pkg.Name)
		return nil, nil, nil
	}

	m, err := registry.PackageInfos.ToMap()
	if err != nil {
		log.New().WithFields(logrus.Fields{
			"registry_name": pkg.Registry,
		}).WithError(err).Warnf("registry is invalid")
		return nil, nil, nil
	}

	pkgInfo, ok := m[pkg.Name]
	if !ok {
		log.New().Warnf("package isn't found %s", pkg.Name)
		return nil, nil, nil
	}
	for _, file := range pkgInfo.GetFiles() {
		if file.Name == exeName {
			return pkg, pkgInfo, file
		}
	}
	return nil, nil, nil
}

func (ctrl *Controller) findExecFile(ctx context.Context, cfgFilePath, exeName string) (*Package, PackageInfo, *File, error) {
	cfg := &Config{}
	if err := ctrl.readConfig(cfgFilePath, cfg); err != nil {
		return nil, nil, nil, err
	}

	registryContents, err := ctrl.installRegistries(ctx, cfg, cfgFilePath)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, pkg := range cfg.Packages {
		if pkg, pkgInfo, file := ctrl.findExecFileFromPkg(registryContents, exeName, pkg); pkg != nil {
			return pkg, pkgInfo, file, nil
		}
	}
	return nil, nil, nil, nil
}

func isUnarchived(archiveType, assetName string) bool {
	return archiveType == "raw" || (archiveType == "" && filepath.Ext(assetName) == "")
}

func (ctrl *Controller) exec(ctx context.Context, pkg *Package, pkgInfo PackageInfo, src string, args []string) error {
	pkgPath, err := pkgInfo.GetPkgPath(ctrl.RootDir, pkg)
	if err != nil {
		return fmt.Errorf("get pkg install path: %w", err)
	}
	exePath := filepath.Join(pkgPath, src)

	if _, err := os.Stat(exePath); err != nil {
		return fmt.Errorf("file.src is invalid. file isn't found %s: %w", exePath, err)
	}

	return ctrl.execCommand(ctx, exePath, args)
}

func (ctrl *Controller) execCommand(ctx context.Context, exePath string, args []string) error {
	cmd := exec.Command(exePath, args...)
	cmd.Stdin = ctrl.Stdin
	cmd.Stdout = ctrl.Stdout
	cmd.Stderr = ctrl.Stderr
	runner := timeout.NewRunner(0)

	logE := log.New().WithField("exe_path", exePath)
	logE.Debug("execute the command")
	if err := runner.Run(ctx, cmd); err != nil {
		exitCode := cmd.ProcessState.ExitCode()
		logE.WithError(err).WithField("exit_code", exitCode).Debug("command was executed but it failed")
		return ecerror.Wrap(err, exitCode)
	}
	return nil
}
