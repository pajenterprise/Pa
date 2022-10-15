// Code generated - DO NOT EDIT.

package eval

type StringOpOverloadBase interface {
	StringEquals(ctx *Context, value string) (bool, error)

	StringNotEquals(ctx *Context, value string) (bool, error)
}

type IntOpOverloadBase interface {
	IntEquals(ctx *Context, value int) (bool, error)

	IntNotEquals(ctx *Context, value int) (bool, error)

	IntAnd(ctx *Context, value int) (int, error)

	IntOr(ctx *Context, value int) (int, error)

	IntXor(ctx *Context, value int) (int, error)

	GreaterThan(ctx *Context, value int) (bool, error)

	GreaterOrEqualThan(ctx *Context, value int) (bool, error)

	LesserThan(ctx *Context, value int) (bool, error)

	LesserOrEqualThan(ctx *Context, value int) (bool, error)
}

type BoolOpOverloadBase interface {
	BoolEquals(ctx *Context, value bool) (bool, error)

	BoolNotEquals(ctx *Context, value bool) (bool, error)
}

func Or(a *BoolEvaluator, b *BoolEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		if state.field != "" {
			if a.isPartial {
				ea = func(ctx *Context) bool {
					return true
				}
			}
			if b.isPartial {
				eb = func(ctx *Context) bool {
					return true
				}
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 || op2
				ctx.Logf("Evaluating %v || %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) || eb(ctx)
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		if state.field != "" {
			if a.isPartial {
				ea = true
			}
			if b.isPartial {
				eb = true
			}
		}

		return &BoolEvaluator{
			Value:     ea || eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		if state.field != "" {
			if a.isPartial {
				ea = func(ctx *Context) bool {
					return true
				}
			}
			if b.isPartial {
				eb = true
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 || op2
				ctx.Logf("Evaluating %v || %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) || eb
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	if state.field != "" {
		if a.isPartial {
			ea = true
		}
		if b.isPartial {
			eb = func(ctx *Context) bool {
				return true
			}
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 || op2
			ctx.Logf("Evaluating %v || %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea || eb(ctx)
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func And(a *BoolEvaluator, b *BoolEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		if state.field != "" {
			if a.isPartial {
				ea = func(ctx *Context) bool {
					return true
				}
			}
			if b.isPartial {
				eb = func(ctx *Context) bool {
					return true
				}
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 && op2
				ctx.Logf("Evaluating %v && %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) && eb(ctx)
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		if state.field != "" {
			if a.isPartial {
				ea = true
			}
			if b.isPartial {
				eb = true
			}
		}

		return &BoolEvaluator{
			Value:     ea && eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		if state.field != "" {
			if a.isPartial {
				ea = func(ctx *Context) bool {
					return true
				}
			}
			if b.isPartial {
				eb = true
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 && op2
				ctx.Logf("Evaluating %v && %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) && eb
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	if state.field != "" {
		if a.isPartial {
			ea = true
		}
		if b.isPartial {
			eb = func(ctx *Context) bool {
				return true
			}
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 && op2
			ctx.Logf("Evaluating %v && %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea && eb(ctx)
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func IntEquals(a *IntEvaluator, b *IntEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 == op2
				ctx.Logf("Evaluating %v == %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) == eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.IntEquals(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.IntEquals(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea == eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 == op2
				ctx.Logf("Evaluating %v == %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) == eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.IntEquals(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 == op2
			ctx.Logf("Evaluating %v == %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea == eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.IntEquals(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func IntNotEquals(a *IntEvaluator, b *IntEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 != op2
				ctx.Logf("Evaluating %v != %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) != eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.IntNotEquals(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.IntNotEquals(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea != eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 != op2
				ctx.Logf("Evaluating %v != %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) != eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.IntNotEquals(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 != op2
			ctx.Logf("Evaluating %v != %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea != eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.IntNotEquals(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func IntAnd(a *IntEvaluator, b *IntEvaluator, opts *Opts, state *state) (*IntEvaluator, error) {
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

		var evalFnc func(ctx *Context) int
		if opts.Debug {
			evalFnc = func(ctx *Context) int {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 & op2
				ctx.Logf("Evaluating %v & %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) int {
				return ea(ctx) & eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) int {
					result, err := a.OpOverload.IntAnd(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) int {
					result, err := b.OpOverload.IntAnd(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &IntEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &IntEvaluator{
			Value:     ea & eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: BitmaskValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) int
		if opts.Debug {
			evalFnc = func(ctx *Context) int {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 & op2
				ctx.Logf("Evaluating %v & %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) int {
				return ea(ctx) & eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) int {
					result, err := a.OpOverload.IntAnd(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &IntEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: BitmaskValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) int
	if opts.Debug {
		evalFnc = func(ctx *Context) int {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 & op2
			ctx.Logf("Evaluating %v & %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) int {
			return ea & eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) int {
				result, err := b.OpOverload.IntAnd(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &IntEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func IntOr(a *IntEvaluator, b *IntEvaluator, opts *Opts, state *state) (*IntEvaluator, error) {
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

		var evalFnc func(ctx *Context) int
		if opts.Debug {
			evalFnc = func(ctx *Context) int {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 | op2
				ctx.Logf("Evaluating %v | %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) int {
				return ea(ctx) | eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) int {
					result, err := a.OpOverload.IntOr(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) int {
					result, err := b.OpOverload.IntOr(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &IntEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &IntEvaluator{
			Value:     ea | eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: BitmaskValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) int
		if opts.Debug {
			evalFnc = func(ctx *Context) int {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 | op2
				ctx.Logf("Evaluating %v | %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) int {
				return ea(ctx) | eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) int {
					result, err := a.OpOverload.IntOr(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &IntEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: BitmaskValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) int
	if opts.Debug {
		evalFnc = func(ctx *Context) int {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 | op2
			ctx.Logf("Evaluating %v | %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) int {
			return ea | eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) int {
				result, err := b.OpOverload.IntOr(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &IntEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func IntXor(a *IntEvaluator, b *IntEvaluator, opts *Opts, state *state) (*IntEvaluator, error) {
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

		var evalFnc func(ctx *Context) int
		if opts.Debug {
			evalFnc = func(ctx *Context) int {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 ^ op2
				ctx.Logf("Evaluating %v ^ %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) int {
				return ea(ctx) ^ eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) int {
					result, err := a.OpOverload.IntXor(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) int {
					result, err := b.OpOverload.IntXor(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &IntEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &IntEvaluator{
			Value:     ea ^ eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: BitmaskValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) int
		if opts.Debug {
			evalFnc = func(ctx *Context) int {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 ^ op2
				ctx.Logf("Evaluating %v ^ %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) int {
				return ea(ctx) ^ eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) int {
					result, err := a.OpOverload.IntXor(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &IntEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: BitmaskValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) int
	if opts.Debug {
		evalFnc = func(ctx *Context) int {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 ^ op2
			ctx.Logf("Evaluating %v ^ %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) int {
			return ea ^ eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) int {
				result, err := b.OpOverload.IntXor(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &IntEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func StringEquals(a *StringEvaluator, b *StringEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 == op2
				ctx.Logf("Evaluating %v == %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) == eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.StringEquals(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.StringEquals(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea == eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 == op2
				ctx.Logf("Evaluating %v == %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) == eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.StringEquals(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 == op2
			ctx.Logf("Evaluating %v == %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea == eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.StringEquals(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func StringNotEquals(a *StringEvaluator, b *StringEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 != op2
				ctx.Logf("Evaluating %v != %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) != eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.StringNotEquals(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.StringNotEquals(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea != eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 != op2
				ctx.Logf("Evaluating %v != %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) != eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.StringNotEquals(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 != op2
			ctx.Logf("Evaluating %v != %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea != eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.StringNotEquals(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func BoolEquals(a *BoolEvaluator, b *BoolEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 == op2
				ctx.Logf("Evaluating %v == %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) == eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.BoolEquals(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.BoolEquals(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea == eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 == op2
				ctx.Logf("Evaluating %v == %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) == eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.BoolEquals(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 == op2
			ctx.Logf("Evaluating %v == %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea == eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.BoolEquals(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func BoolNotEquals(a *BoolEvaluator, b *BoolEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 != op2
				ctx.Logf("Evaluating %v != %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) != eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.BoolNotEquals(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.BoolNotEquals(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea != eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 != op2
				ctx.Logf("Evaluating %v != %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) != eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.BoolNotEquals(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 != op2
			ctx.Logf("Evaluating %v != %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea != eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.BoolNotEquals(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func GreaterThan(a *IntEvaluator, b *IntEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 > op2
				ctx.Logf("Evaluating %v > %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) > eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.GreaterThan(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.GreaterThan(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea > eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 > op2
				ctx.Logf("Evaluating %v > %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) > eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.GreaterThan(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 > op2
			ctx.Logf("Evaluating %v > %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea > eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.GreaterThan(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func GreaterOrEqualThan(a *IntEvaluator, b *IntEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 >= op2
				ctx.Logf("Evaluating %v >= %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) >= eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.GreaterOrEqualThan(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.GreaterOrEqualThan(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea >= eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 >= op2
				ctx.Logf("Evaluating %v >= %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) >= eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.GreaterOrEqualThan(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 >= op2
			ctx.Logf("Evaluating %v >= %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea >= eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.GreaterOrEqualThan(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func LesserThan(a *IntEvaluator, b *IntEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 < op2
				ctx.Logf("Evaluating %v < %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) < eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.LesserThan(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.LesserThan(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea < eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 < op2
				ctx.Logf("Evaluating %v < %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) < eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.LesserThan(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 < op2
			ctx.Logf("Evaluating %v < %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea < eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.LesserThan(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}

func LesserOrEqualThan(a *IntEvaluator, b *IntEvaluator, opts *Opts, state *state) (*BoolEvaluator, error) {
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

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb(ctx)
				result := op1 <= op2
				ctx.Logf("Evaluating %v <= %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) <= eb(ctx)
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.LesserOrEqualThan(ctx, eb(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			} else if b.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := b.OpOverload.LesserOrEqualThan(ctx, ea(ctx))
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc == nil && b.EvalFnc == nil {
		ea, eb := a.Value, b.Value

		return &BoolEvaluator{
			Value:     ea <= eb,
			isPartial: isPartialLeaf,
		}, nil
	}

	if a.EvalFnc != nil || a.OpOverload != nil {
		ea, eb := a.EvalFnc, b.Value

		if a.Field != "" {
			if err := state.UpdateFieldValues(a.Field, FieldValue{Value: eb, Type: ScalarValueType}); err != nil {
				return nil, err
			}
		}

		var evalFnc func(ctx *Context) bool
		if opts.Debug {
			evalFnc = func(ctx *Context) bool {
				ctx.evalDepth++
				op1, op2 := ea(ctx), eb
				result := op1 <= op2
				ctx.Logf("Evaluating %v <= %v => %v", op1, op2, result)
				ctx.evalDepth--
				return result
			}
		} else {
			evalFnc = func(ctx *Context) bool {
				return ea(ctx) <= eb
			}

			if a.OpOverload != nil {
				evalFnc = func(ctx *Context) bool {
					result, err := a.OpOverload.LesserOrEqualThan(ctx, eb)
					if err != nil {
						return evalFnc(ctx)
					}
					return result
				}
			}

		}

		return &BoolEvaluator{
			EvalFnc:   evalFnc,
			isPartial: isPartialLeaf,
		}, nil
	}

	ea, eb := a.Value, b.EvalFnc

	if b.Field != "" {
		if err := state.UpdateFieldValues(b.Field, FieldValue{Value: ea, Type: ScalarValueType}); err != nil {
			return nil, err
		}
	}

	var evalFnc func(ctx *Context) bool
	if opts.Debug {
		evalFnc = func(ctx *Context) bool {
			ctx.evalDepth++
			op1, op2 := ea, eb(ctx)
			result := op1 <= op2
			ctx.Logf("Evaluating %v <= %v => %v", op1, op2, result)
			ctx.evalDepth--
			return result
		}
	} else {
		evalFnc = func(ctx *Context) bool {
			return ea <= eb(ctx)
		}

		if a.OpOverload != nil {
			evalFnc = func(ctx *Context) bool {
				result, err := b.OpOverload.LesserOrEqualThan(ctx, ea)
				if err != nil {
					return evalFnc(ctx)
				}
				return result
			}
		}

	}

	return &BoolEvaluator{
		EvalFnc:   evalFnc,
		isPartial: isPartialLeaf,
	}, nil
}
