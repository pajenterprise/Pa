// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package main

import (
	"flag"
	"os"
	"os/exec"
	"text/template"
)

var (
	output string
)

func main() {
	tmpl := template.Must(template.New("header").Parse(`

// Code generated - DO NOT EDIT.

package	eval

type StringOpOverloadBase interface {
{{ range . }}
	{{ if and (eq .ArgsType "StringEvaluator") .Overloadable }}
	{{ .FuncName }}(ctx *Context, value string) ({{ .EvalReturnType }}, error)
	{{ end }}
{{ end }}
}

type IntOpOverloadBase interface {
{{ range . }}
	{{ if and (eq .ArgsType "IntEvaluator") .Overloadable }}
	{{ .FuncName }}(ctx *Context, value int) ({{ .EvalReturnType }}, error)
	{{ end }}
{{ end }}
}

type BoolOpOverloadBase interface {
{{ range . }}
	{{ if and (eq .ArgsType "BoolEvaluator") .Overloadable }}
	{{ .FuncName }}(ctx *Context, value bool) ({{ .EvalReturnType }}, error)
	{{ end }}
{{ end }}
}

{{ range . }}

func {{ .FuncName }}(a *{{ .ArgsType }}, b *{{ .ArgsType }}, opts *Opts, state *state) (*{{ .FuncReturnType }}, error) {
	partialA, partialB := a.isPartial, b.isPartial

	if a.EvalFnc == nil || (a.Field != "" && a.Field != state.field) {
		partialA = true
	}
	if b.EvalFnc == nil || (b.Field != "" && b.Field != state.field) {
		partialB = true
	}
	isPartialLeaf := partialA && partialB

	if a.Field != "" && b.Field != "" {
		isPartialLeaf = true
	}

	if a.EvalFnc != nil && b.EvalFnc != nil {
		ea, eb := a.EvalFnc, b.EvalFnc

		{{ if or (eq .FuncName "Or") (eq .FuncName "And") }}
			if state.field != "" {
				if a.isPartial {
					ea = func(ctx *Context) {{ .EvalReturnType }} {
						return true
					}
				}
				if b.isPartial {
					eb = func(ctx *Context) {{ .EvalReturnType }} {
						return true
					}
				}
			}
		{{ end }}

		var evalFnc func(ctx *Context) {{ .EvalReturnType }}
		if opts.Debug {
			evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 {{.Op}} op2
				ctx.Logf("Evaluating %v {{ .Op }} %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
				return ea(ctx) {{ .Op }} eb(ctx)
			}
			{{ if .Overloadable }}
				if a.OpOverload != nil {
					evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
						result, err := a.OpOverload.{{ .FuncName }}(ctx, eb(ctx))
						if err != nil {
							return evalFnc(ctx)
						} 
						return result
					}
				} else if b.OpOverload != nil {
					evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
						result, err := b.OpOverload.{{ .FuncName }}(ctx, ea(ctx))
						if err != nil {
							return evalFnc(ctx)
						}
						return result
					}
				}
			{{ end }}	
		}

		return &{{ .FuncReturnType }}{
			EvalFnc: evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		{{ if or (eq .FuncName "Or") (eq .FuncName "And") }}
		if state.field != "" {
			if a.isPartial {
				ea = true
			}
			if b.isPartial {
				eb = true
			}
		}
		{{ end }}

		return &{{ .FuncReturnType }}{
			Value: ea {{ .Op }} eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: {{ .ValueType }}}); err != nil {
				return nil, err
			}
		}

		{{ if or (eq .FuncName "Or") (eq .FuncName "And") }}
			if state.field != "" {
				if a.isPartial {
					ea = func(ctx *Context) {{ .EvalReturnType }} {
						return true
					}
				}
				if b.isPartial {
					eb = true
				}
			}
		{{ end }}

		var evalFnc func(ctx *Context) {{ .EvalReturnType }}
		if opts.Debug {
			evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 {{ .Op }} op2
				ctx.Logf("Evaluating %v {{.Op}} %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
				return ea(ctx) {{ .Op }} eb
			}
			{{ if .Overloadable }} 
				if a.OpOverload != nil {
					evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
						result, err := a.OpOverload.{{ .FuncName }}(ctx, eb)
						if err != nil {
							return evalFnc(ctx)
						}
						return result
					}
				}
			{{ end }}
		}

		return &{{ .FuncReturnType }}{
			EvalFnc: evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: {{ .ValueType }}}); err != nil {
			return nil, err
		}
	}

	{{ if or (eq .FuncName "Or") (eq .FuncName "And") }}
		if state.field != "" {
			if a.isPartial {
				ea = true
			}
			if b.isPartial {
				eb = func(ctx *Context) {{ .EvalReturnType }} {
					return true
				}
			}
		}
	{{ end }}

	var evalFnc func(ctx *Context) {{ .EvalReturnType }}
	if opts.Debug {
		evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 {{ .Op }} op2
			ctx.Logf("Evaluating %v {{ .Op }} %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
			return ea {{ .Op }} eb(ctx)
		}
		{{ if .Overloadable }} 
			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) {{ .EvalReturnType }} {
					result, err := b.OpOverload.{{ .FuncName }}(ctx, ea)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}
		{{ end }}
	}

	return &{{ .FuncReturnType }}{
		EvalFnc: evalFnc	,
		isPartial: isPartialLeaf,
	}, nil
}
{{ end }}
`))

	outputFile, err := os.Create(output)
	if err != nil {
		panic(err)
	}

	operators := []struct {
		FuncName       string
		ArgsType       string
		FuncReturnType string
		EvalReturnType string
		Op             string
		ValueType      string
		Overloadable   bool
	}{
		{
			FuncName:       "Or",
			ArgsType:       "BoolEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "||",
			ValueType:      "ScalarValueType",
		},
		{
			FuncName:       "And",
			ArgsType:       "BoolEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "&&",
			ValueType:      "ScalarValueType",
		},
		{
			FuncName:       "IntEquals",
			ArgsType:       "IntEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "==",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "IntNotEquals",
			ArgsType:       "IntEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "!=",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "IntAnd",
			ArgsType:       "IntEvaluator",
			FuncReturnType: "IntEvaluator",
			EvalReturnType: "int",
			Op:             "&",
			ValueType:      "BitmaskValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "IntOr",
			ArgsType:       "IntEvaluator",
			FuncReturnType: "IntEvaluator",
			EvalReturnType: "int",
			Op:             "|",
			ValueType:      "BitmaskValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "IntXor",
			ArgsType:       "IntEvaluator",
			FuncReturnType: "IntEvaluator",
			EvalReturnType: "int",
			Op:             "^",
			ValueType:      "BitmaskValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "StringEquals",
			ArgsType:       "StringEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "==",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "StringNotEquals",
			ArgsType:       "StringEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "!=",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "BoolEquals",
			ArgsType:       "BoolEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "==",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "BoolNotEquals",
			ArgsType:       "BoolEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "!=",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "GreaterThan",
			ArgsType:       "IntEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             ">",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "GreaterOrEqualThan",
			ArgsType:       "IntEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             ">=",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "LesserThan",
			ArgsType:       "IntEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "<",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
		{
			FuncName:       "LesserOrEqualThan",
			ArgsType:       "IntEvaluator",
			FuncReturnType: "BoolEvaluator",
			EvalReturnType: "bool",
			Op:             "<=",
			ValueType:      "ScalarValueType",
			Overloadable:   true,
		},
	}

	if err := tmpl.Execute(outputFile, operators); err != nil {
		panic(err)
	}

	if err := outputFile.Close(); err != nil {
		panic(err)
	}

	cmd := exec.Command("gofmt", "-s", "-w", output)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func init() {
	flag.StringVar(&output, "output", "", "Go generated file")
	flag.Parse()
}
