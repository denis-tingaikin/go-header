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

//Scope means the scope for go-header linter in project
type Scope struct {
	//Policy means file scoe policy. Can be "none", "diff", "new".
	Policy GitPolicy `yaml:"policy"`
	//MasterBranch master branch for scope. Used only if Policy is not "none".
	MasterBranch string `yaml:"master-branch"`
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
