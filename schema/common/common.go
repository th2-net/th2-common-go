package common

type Module interface {
	GetKey() ModuleKey
}

type ModuleKey string
