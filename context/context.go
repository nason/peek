package context

import (
	"errors"

	"peek/git"
)

// GetRemotes for the current git context
func GetRemotes() (Remotes, error) {
	gitRemotes, err := git.Remotes()
	if err != nil {
		return nil, err
	}
	if len(gitRemotes) == 0 {
		return nil, errors.New("no git remotes found")
	}

	sshTranslate := git.ParseSSHConfig().Translator()
	remotes := translateRemotes(gitRemotes, sshTranslate)

	return remotes, nil
}
