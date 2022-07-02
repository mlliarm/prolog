package engine

import (
	"io"
)

// Term is a prolog term.
type Term interface {
	Unify(Term, bool, *Env) (*Env, bool)
	WriteTerm(w io.Writer, opts *WriteOptions, env *Env) error
	Compare(Term, *Env) int64
}

// Contains checks if t contains s.
func Contains(t, s Term, env *Env) bool {
	switch t := t.(type) {
	case Variable:
		if t == s {
			return true
		}
		ref, ok := env.Lookup(t)
		if !ok {
			return false
		}
		return Contains(ref, s, env)
	case *Compound:
		if s, ok := s.(Atom); ok && t.Functor == s {
			return true
		}
		for _, a := range t.Args {
			if Contains(a, s, env) {
				return true
			}
		}
		return false
	default:
		return t == s
	}
}

// Rulify returns t if t is in a form of P:-Q, t:-true otherwise.
func Rulify(t Term, env *Env) Term {
	t = env.Resolve(t)
	if c, ok := t.(*Compound); ok && c.Functor == ":-" && len(c.Args) == 2 {
		return t
	}
	return &Compound{Functor: ":-", Args: []Term{t, Atom("true")}}
}

// WriteOptions specify how the Term writes itself.
type WriteOptions struct {
	IgnoreOps     bool
	Quoted        bool
	VariableNames map[Variable]Atom
	NumberVars    bool

	ops           operators
	priority      Integer
	visited       map[Term]struct{}
	prefixMinus   bool
	before, after operator
}

func (o WriteOptions) withQuoted(quoted bool) *WriteOptions {
	o.Quoted = quoted
	return &o
}

func (o WriteOptions) withFreshVisited() *WriteOptions {
	visited := make(map[Term]struct{}, len(o.visited))
	for k, v := range o.visited {
		visited[k] = v
	}
	o.visited = visited
	return &o
}

func (o WriteOptions) withPriority(priority Integer) *WriteOptions {
	o.priority = priority
	return &o
}

func (o WriteOptions) withBefore(op operator) *WriteOptions {
	o.before = op
	return &o
}

func (o WriteOptions) withAfter(op operator) *WriteOptions {
	o.after = op
	return &o
}

var defaultWriteOptions = WriteOptions{
	ops: operators{
		{priority: 500, specifier: operatorSpecifierYFX, name: "+"}, // for flag+value
		{priority: 400, specifier: operatorSpecifierYFX, name: "/"}, // for principal functors
	},
	VariableNames: map[Variable]Atom{},
	priority:      1200,
}

// https://go.dev/blog/errors-are-values
type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) Write(p []byte) (int, error) {
	if ew.err != nil {
		return 0, nil
	}
	var n int
	n, ew.err = ew.w.Write(p)
	return n, nil
}
