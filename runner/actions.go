package runner

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// FilesPlaceholder is the argument of an action that is replaced by the paths
// the user passed.
const FilesPlaceholder = "$FILES"

var actionName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// ErrActionTakesNoFiles is returned when a user passes paths to an action whose
// definition has no place for them.
var ErrActionTakesNoFiles = errors.New("this command does not take any arguments")

// ValidateAction checks the name and definition of an action.
//
// An action is the only thing a user of the command runner can start. Its
// definition is a program followed by fixed arguments, optionally with one
// $FILES argument marking where the user's paths go. The program must be an
// absolute path so that neither PATH nor the working directory, which is
// inside the user's scope, decides which binary runs.
func ValidateAction(name, definition string) error {
	if !actionName.MatchString(name) {
		return fmt.Errorf("invalid name %q: use letters, digits, '.', '_' and '-'", name)
	}

	program, args, err := SplitCommandAndArgs(definition)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(program) {
		return fmt.Errorf("the program must be an absolute path, got %q", program)
	}

	placeholders := 0
	for _, arg := range args {
		switch {
		case arg == FilesPlaceholder:
			placeholders++
		case strings.Contains(arg, FilesPlaceholder):
			return fmt.Errorf("%s must be an argument on its own, got %q", FilesPlaceholder, arg)
		}
	}
	if placeholders > 1 {
		return fmt.Errorf("%s can only be used once", FilesPlaceholder)
	}

	return nil
}

// BuildAction returns the program and arguments to run for an action, with
// paths put in place of $FILES.
//
// Nothing the user typed reaches the result except through paths, and the
// caller must have resolved each of those to a real path inside the user's
// scope. A path that would be read as an option is refused.
func BuildAction(name, definition string, paths []string) (program string, args []string, err error) {
	if err := ValidateAction(name, definition); err != nil {
		return "", nil, err
	}

	program, template, err := SplitCommandAndArgs(definition)
	if err != nil {
		return "", nil, err
	}

	for _, p := range paths {
		if !filepath.IsAbs(p) || strings.HasPrefix(p, "-") {
			return "", nil, fmt.Errorf("refusing to pass %q as an argument", p)
		}
	}

	takesFiles := false
	for _, arg := range template {
		if arg == FilesPlaceholder {
			takesFiles = true
			args = append(args, paths...)
			continue
		}
		args = append(args, arg)
	}

	if !takesFiles && len(paths) > 0 {
		return "", nil, ErrActionTakesNoFiles
	}

	return program, args, nil
}
