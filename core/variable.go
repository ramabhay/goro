package core

import (
	"io"

	"github.com/MagicalTux/gophp/core/tokenizer"
)

type runVariable struct {
	v ZString
	l *Loc
}

type runVariableRef struct {
	v Runnable
	l *Loc
}

func compileRunVariableRef(i *tokenizer.Item, c compileCtx, l *Loc) (Runnable, error) {
	r := &runVariableRef{l: l}
	var err error

	if i == nil {
		i, err = c.NextItem()
		if err != nil {
			return nil, err
		}
	}

	if i.Type == tokenizer.Rune('{') {
		r.v, err = compileExpr(nil, c)
		if err != nil {
			return nil, err
		}

		i, err = c.NextItem()
		if err != nil {
			return nil, err
		}
		if i.Type != tokenizer.Rune('}') {
			return nil, i.Unexpected()
		}
	} else {
		r.v, err = compileOneExpr(i, c)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *runVariable) Run(ctx Context) (*ZVal, error) {
	res, err := ctx.OffsetGet(ctx, r.v.ZVal())
	return res.Nude(), err
}

func (r *runVariable) WriteValue(ctx Context, value *ZVal) error {
	var err error
	if value == nil {
		err = ctx.OffsetUnset(ctx, r.v.ZVal())
	} else {
		err = ctx.OffsetSet(ctx, r.v.ZVal(), value)
	}
	if err != nil {
		return r.l.Error(err)
	}
	return nil
}

func (r *runVariable) Loc() *Loc {
	return r.l
}

func (r *runVariable) Dump(w io.Writer) error {
	_, err := w.Write([]byte{'$'})
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(r.v))
	return err
}

func (r *runVariableRef) Dump(w io.Writer) error {
	_, err := w.Write([]byte("${"))
	if err != nil {
		return err
	}
	err = r.v.Dump(w)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte{'}'})
	return err
}

func (r *runVariableRef) Loc() *Loc {
	return r.l
}

func (r *runVariableRef) Run(ctx Context) (*ZVal, error) {
	v, err := r.v.Run(ctx)
	if err != nil {
		return nil, err
	}
	v, err = ctx.OffsetGet(ctx, v)
	if v != nil {
		v = v.Nude()
	}
	return v, err
}

func (r *runVariableRef) WriteValue(ctx Context, value *ZVal) error {
	var err error
	v, err := r.v.Run(ctx)
	if err != nil {
		return err
	}

	if value == nil {
		err = ctx.OffsetUnset(ctx, v)
	} else {
		err = ctx.OffsetSet(ctx, v, value)
	}
	if err != nil {
		return r.l.Error(err)
	}
	return nil
}

// reference to an existing [something]
type runRef struct {
	v Runnable
	l *Loc
}

func (r *runRef) Loc() *Loc {
	return r.l
}

func (r *runRef) Run(ctx Context) (*ZVal, error) {
	z, err := r.v.Run(ctx)
	if err != nil {
		return nil, err
	}
	// embed zval into another zval
	return z.Ref(), nil
}

func (r *runRef) Dump(w io.Writer) error {
	_, err := w.Write([]byte{'&'})
	if err != nil {
		return err
	}
	return r.v.Dump(w)
}
