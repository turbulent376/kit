package config

import (
	errors "git.jetbrains.space/orbi/fcsd/kit/er"
)

var (
	ErrConfigTargetObjectInvalidType = func() error {
		return errors.WithBuilder("CFG-001", "config target is not correct format").Err()
	}
	ErrConfigPaErrConfigPathIsEmpty = func() error {
		return errors.WithBuilder("CFG-002", "config path is empty").Err()
	}
	ErrConfigFileNotFound = func(v string) error {
		return errors.WithBuilder("CFG-003", "config file on path %s is not found", v).Err()
	}
	ErrConfigFileOpen = func(err error, v string) error {
		return errors.WrapWithBuilder(err, "CFG-004", "open file: %s", v).Err()
	}
	ErrConfigInit = func(err error) error {
		return errors.WrapWithBuilder(err, "CFG-005", "can not init config").Err()
	}
	ErrConfigLoad = func(err error) error {
		return errors.WrapWithBuilder(err, "CFG-006", "can not load config").Err()
	}
)
