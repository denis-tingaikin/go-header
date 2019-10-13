package models

import (
	"github.com/denis-tingajkin/go-header/messages"
)

type GitPolicy string

const (
	DiffOnlyPolicy     GitPolicy = "diff"
	NonePolicy         GitPolicy = "none"
	OnlyNewFilesPolicy GitPolicy = "new"
)

type Scope struct {
	Policy       GitPolicy `yaml:"policy"`
	MasterBranch string    `yaml:"master-branch"`
}

func (s Scope) Validate() error {
	if s.Policy != NonePolicy && s.Policy != DiffOnlyPolicy && s.Policy != OnlyNewFilesPolicy {
		return messages.UnknownField(string(s.Policy))
	}
	if s.Policy != NonePolicy && s.MasterBranch == "" {
		return messages.MasterBranchCanNotBeEmpty()
	}
	return nil
}
