package papercuts

import (
	"fmt"
	"os"
	"path/filepath"
)

// Scope selects a project or global papercuts log.
type Scope uint8

const (
	// ProjectScope selects PAPERCUTS.md in the invocation working directory.
	ProjectScope Scope = iota + 1
	// GlobalScope selects the configured user-global log.
	GlobalScope
)

// String returns the CLI name of the scope.
func (s Scope) String() string {
	switch s {
	case ProjectScope:
		return "project"
	case GlobalScope:
		return "global"
	default:
		return ""
	}
}

// TargetOptions contains scope flags preserved from CLI parsing.
type TargetOptions struct {
	Project    bool
	Global     bool
	GlobalPath *string
}

type resolvedTarget struct {
	scope            Scope
	logPath          string
	agentsPath       string
	globalDirs       bool
	customGlobalPath bool
}

type systemSources struct {
	getwd       func() (string, error)
	lookupEnv   func(string) (string, bool)
	userHomeDir func() (string, error)
}

func defaultSystemSources() systemSources {
	return systemSources{
		getwd:       os.Getwd,
		lookupEnv:   os.LookupEnv,
		userHomeDir: os.UserHomeDir,
	}
}

func resolveTarget(options TargetOptions, sources systemSources) (resolvedTarget, error) {
	workingDirectory, err := sources.getwd()
	if err != nil {
		return resolvedTarget{}, fmt.Errorf("resolve working directory: %w", err)
	}
	if options.Project && options.Global {
		return resolvedTarget{}, &ValidationError{Field: "scope", Reason: "cannot select both project and global"}
	}

	global := options.Global
	if options.GlobalPath != nil && !global {
		return resolvedTarget{}, &ValidationError{Field: "global path", Reason: "requires global scope"}
	}
	if !global {
		return resolvedTarget{
			scope:      ProjectScope,
			logPath:    filepath.Join(workingDirectory, "PAPERCUTS.md"),
			agentsPath: filepath.Join(workingDirectory, "AGENTS.md"),
		}, nil
	}

	globalPath := ""
	customGlobalPath := false
	if options.GlobalPath != nil {
		customGlobalPath = true
		globalPath = *options.GlobalPath
		if globalPath == "" {
			return resolvedTarget{}, &ValidationError{Field: "global path", Reason: "must not be empty"}
		}
	} else if environmentPath, ok := sources.lookupEnv("PAPERCUTS_GLOBAL_PATH"); ok && environmentPath != "" {
		customGlobalPath = true
		globalPath = environmentPath
	} else {
		home, homeErr := sources.userHomeDir()
		if homeErr != nil {
			return resolvedTarget{}, fmt.Errorf("resolve user home: %w", homeErr)
		}
		if !filepath.IsAbs(home) {
			return resolvedTarget{}, &ValidationError{Field: "user home", Reason: "must be absolute"}
		}
		globalPath = filepath.Join(home, ".papercuts", "PAPERCUTS.md")
	}
	if !filepath.IsAbs(globalPath) {
		return resolvedTarget{}, &ValidationError{Field: "global path", Reason: fmt.Sprintf("must be absolute: %q", globalPath)}
	}

	return resolvedTarget{
		scope:            GlobalScope,
		logPath:          filepath.Clean(globalPath),
		agentsPath:       filepath.Join(workingDirectory, "AGENTS.md"),
		globalDirs:       true,
		customGlobalPath: customGlobalPath,
	}, nil
}
