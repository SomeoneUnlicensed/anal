package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/SomeoneUnlicensed/anal/skills"
)

func runInstall(args []string) error {
	target := "all"
	if len(args) > 0 {
		target = args[0]
	}

	var dirs []string
	switch target {
	case "codex":
		dir, err := codexSkillsDir()
		if err != nil {
			return err
		}
		dirs = []string{dir}
	case "claude":
		dir, err := claudeSkillsDir()
		if err != nil {
			return err
		}
		dirs = []string{dir}
	case "all":
		codexDir, err := codexSkillsDir()
		if err != nil {
			return err
		}
		claudeDir, err := claudeSkillsDir()
		if err != nil {
			return err
		}
		dirs = []string{codexDir, claudeDir}
	default:
		return fmt.Errorf("unknown install target %q (use codex, claude, or all)", target)
	}

	for _, dir := range dirs {
		if err := writeSkill(dir); err != nil {
			return err
		}
		fmt.Printf("installed ANAL skill to %s\n", dir)
	}
	return nil
}

func codexSkillsDir() (string, error) {
	home := os.Getenv("CODEX_HOME")
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		home = filepath.Join(userHome, ".codex")
	}
	return filepath.Join(home, "skills"), nil
}

func claudeSkillsDir() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHome, ".claude", "skills"), nil
}

func writeSkill(skillsDir string) error {
	root := filepath.Join(skillsDir, "anal")
	return fs.WalkDir(skill.FS, "anal", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, "anal/")
		dest := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		data, err := fs.ReadFile(skill.FS, path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
}
