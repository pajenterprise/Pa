// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// +build linux

package probe

import (
	"github.com/DataDog/datadog-agent/pkg/security/secl/eval"
)

type PathnameOpOverload struct {
	GetInode func(ctx *eval.Context) uint64
}

func (o *PathnameOpOverload) StringArrayContains(ctx *eval.Context, value []string) (bool, error) {
	return false, eval.ErrOpOverloadNotSupported
}

func (o *PathnameOpOverload) StringEquals(ctx *eval.Context, value string) (bool, error) {
	//return (*Event)(ctx.Object).unlink.crc32 == crc32.ChecksumIEEE([]byte(value)), nil
}

func (o *PathnameOpOverload) StringNotEquals(ctx *eval.Context, value string) (bool, error) {
	result, _ := o.StringEquals(ctx, value)
	return !result, nil
}

func (o *PathnameOpOverload) StringMatches(ctx *eval.Context, value string) (bool, error) {
	return false, eval.ErrOpOverloadNotSupported
}

func NewPathnameOpOverload(field eval.Field) eval.StringOpOverload {
	var getter func(ctx *eval.Context) uint64

	switch field {
	case "open.filename":
		getter = func(ctx *eval.Context) uint64 {
			return (*Event)(ctx.Object).Open.Inode
		}
	default:
		return nil
	}

	return &PathnameOpOverload{
		GetInode: getter,
	}
}
