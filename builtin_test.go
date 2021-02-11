package prolog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

func TestCopyTerm(t *testing.T) {
	in := &Variable{Ref: Atom("a")}
	out := &Variable{}
	k := func() (bool, error) {
		return true, nil
	}
	ok, err := CopyTerm(in, out, k)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, Atom("a"), out.Ref)
}

func TestRepeat(t *testing.T) {
	c := 3
	ok, err := Repeat(func() (bool, error) {
		c--
		return c == 0, nil
	})
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Repeat(func() (bool, error) {
		return false, errors.New("")
	})
	assert.Error(t, err)
	assert.False(t, ok)

	ok, err = Repeat(func() (bool, error) {
		return true, errCut
	})
	assert.True(t, errors.Is(err, errCut))
	assert.True(t, ok)
}

func TestBagOf(t *testing.T) {
	e, err := NewEngine(nil, nil)
	assert.NoError(t, err)
	assert.NoError(t, e.Load(`
foo(a, b, c).
foo(a, b, d).
foo(b, c, e).
foo(b, c, f).
foo(c, c, g).
`))

	t.Run("without qualifier", func(t *testing.T) {
		var c int
		ok, err := e.Query(`bagof(C, foo(A, B, C), Cs).`, func(vs []*Variable) bool {
			switch c {
			case 0:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: Atom("a")},
					{Name: "B", Ref: Atom("b")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("c")}},
							&Variable{Ref: &Variable{Ref: Atom("d")}},
						),
					}},
				}, vs)
			case 1:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: Atom("b")},
					{Name: "B", Ref: Atom("c")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("e")}},
							&Variable{Ref: &Variable{Ref: Atom("f")}},
						),
					}},
				}, vs)
			case 2:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: Atom("c")},
					{Name: "B", Ref: Atom("c")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("g")}},
						),
					}},
				}, vs)
			default:
				assert.Fail(t, "unreachable")
			}
			c++
			return false
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("with qualifier", func(t *testing.T) {
		var c int
		ok, err := e.Query(`bagof(C, A^foo(A, B, C), Cs).`, func(vs []*Variable) bool {
			switch c {
			case 0:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: &Variable{}},
					{Name: "B", Ref: Atom("b")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("c")}},
							&Variable{Ref: &Variable{Ref: Atom("d")}},
						),
					}},
				}, vs)
			case 1:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: &Variable{}},
					{Name: "B", Ref: Atom("c")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("e")}},
							&Variable{Ref: &Variable{Ref: Atom("f")}},
							&Variable{Ref: &Variable{Ref: Atom("g")}},
						),
					}},
				}, vs)
			default:
				assert.Fail(t, "unreachable")
			}
			c++
			return false
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("with multiple qualifiers", func(t *testing.T) {
		var c int
		ok, err := e.Query(`bagof(C, (A, B)^foo(A, B, C), Cs).`, func(vs []*Variable) bool {
			switch c {
			case 0:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A"},
					{Name: "B"},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("c")}},
							&Variable{Ref: &Variable{Ref: Atom("d")}},
							&Variable{Ref: &Variable{Ref: Atom("e")}},
							&Variable{Ref: &Variable{Ref: Atom("f")}},
							&Variable{Ref: &Variable{Ref: Atom("g")}},
						),
					}},
				}, vs)
			default:
				assert.Fail(t, "unreachable")
			}
			c++
			return false
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestSetOf(t *testing.T) {
	e, err := NewEngine(nil, nil)
	assert.NoError(t, err)
	assert.NoError(t, e.Load(`
foo(a, b, c).
foo(a, b, d).
foo(a, b, c).
foo(b, c, e).
foo(b, c, f).
foo(b, c, e).
foo(c, c, g).
foo(c, c, g).
`))

	t.Run("without qualifier", func(t *testing.T) {
		var c int
		ok, err := e.Query(`setof(C, foo(A, B, C), Cs).`, func(vs []*Variable) bool {
			switch c {
			case 0:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: Atom("a")},
					{Name: "B", Ref: Atom("b")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("c")}},
							&Variable{Ref: &Variable{Ref: Atom("d")}},
						),
					}},
				}, vs)
			case 1:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: Atom("b")},
					{Name: "B", Ref: Atom("c")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("e")}},
							&Variable{Ref: &Variable{Ref: Atom("f")}},
						),
					}},
				}, vs)
			case 2:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: Atom("c")},
					{Name: "B", Ref: Atom("c")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("g")}},
						),
					}},
				}, vs)
			default:
				assert.Fail(t, "unreachable")
			}
			c++
			return false
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("with qualifier", func(t *testing.T) {
		var c int
		ok, err := e.Query(`setof(C, A^foo(A, B, C), Cs).`, func(vs []*Variable) bool {
			switch c {
			case 0:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: &Variable{}},
					{Name: "B", Ref: Atom("b")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("c")}},
							&Variable{Ref: &Variable{Ref: Atom("d")}},
						),
					}},
				}, vs)
			case 1:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A", Ref: &Variable{}},
					{Name: "B", Ref: Atom("c")},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("e")}},
							&Variable{Ref: &Variable{Ref: Atom("f")}},
							&Variable{Ref: &Variable{Ref: Atom("g")}},
						),
					}},
				}, vs)
			default:
				assert.Fail(t, "unreachable")
			}
			c++
			return false
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("with multiple qualifiers", func(t *testing.T) {
		var c int
		ok, err := e.Query(`setof(C, (A, B)^foo(A, B, C), Cs).`, func(vs []*Variable) bool {
			switch c {
			case 0:
				assert.Equal(t, []*Variable{
					{Name: "C", Ref: &Variable{}},
					{Name: "A"},
					{Name: "B"},
					{Name: "Cs", Ref: &Variable{
						Ref: List(
							&Variable{Ref: &Variable{Ref: Atom("c")}},
							&Variable{Ref: &Variable{Ref: Atom("d")}},
							&Variable{Ref: &Variable{Ref: Atom("e")}},
							&Variable{Ref: &Variable{Ref: Atom("f")}},
							&Variable{Ref: &Variable{Ref: Atom("g")}},
						),
					}},
				}, vs)
			default:
				assert.Fail(t, "unreachable")
			}
			c++
			return false
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestCompare(t *testing.T) {
	var vs [2]Variable
	ok, err := Compare(Atom("<"), &vs[0], &vs[1], Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("="), &vs[0], &vs[0], Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), &vs[1], &vs[0], Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	vs[0].Ref = Atom("b")
	vs[1].Ref = Atom("a")
	ok, err = Compare(Atom(">"), &vs[0], &vs[1], Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), &Variable{}, Integer(0), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), &Variable{}, Atom(""), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), &Variable{}, &Compound{}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), Integer(0), &Variable{}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), Integer(0), Integer(1), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("="), Integer(0), Integer(0), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), Integer(1), Integer(0), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), Integer(0), Atom(""), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), Integer(0), &Compound{}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), Atom(""), &Variable{}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), Atom(""), Integer(0), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), Atom("a"), Atom("b"), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("="), Atom("a"), Atom("a"), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), Atom("b"), Atom("a"), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), Atom(""), &Compound{}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), &Compound{}, &Variable{}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), &Compound{}, Integer(0), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), &Compound{}, Atom(""), Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), &Compound{Functor: "a"}, &Compound{Functor: "b"}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("="), &Compound{Functor: "a"}, &Compound{Functor: "a"}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), &Compound{Functor: "b"}, &Compound{Functor: "a"}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), &Compound{Functor: "f", Args: []Term{Atom("a")}}, &Compound{Functor: "f"}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("="), &Compound{Functor: "f", Args: []Term{Atom("a")}}, &Compound{Functor: "f", Args: []Term{Atom("a")}}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), &Compound{Functor: "f"}, &Compound{Functor: "f", Args: []Term{Atom("a")}}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom(">"), &Compound{Functor: "f", Args: []Term{Atom("b")}}, &Compound{Functor: "f", Args: []Term{Atom("a")}}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = Compare(Atom("<"), &Compound{Functor: "f", Args: []Term{Atom("a")}}, &Compound{Functor: "f", Args: []Term{Atom("b")}}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestThrow(t *testing.T) {
	ok, err := Throw(Atom("a"), Done)
	assert.Equal(t, &Exception{Term: Atom("a")}, err)
	assert.False(t, ok)
}

func TestEngine_Catch(t *testing.T) {
	e, err := NewEngine(nil, nil)
	assert.NoError(t, err)

	t.Run("match", func(t *testing.T) {
		var v Variable
		ok, err := e.Catch(&Compound{
			Functor: "throw",
			Args:    []Term{Atom("a")},
		}, &v, &Compound{
			Functor: "=",
			Args:    []Term{&v, Atom("a")},
		}, Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("not match", func(t *testing.T) {
		ok, err := e.Catch(&Compound{
			Functor: "throw",
			Args:    []Term{Atom("a")},
		}, Atom("b"), Atom("fail"), Done)
		assert.Equal(t, &Exception{Term: Atom("a")}, err)
		assert.False(t, ok)
	})

	t.Run("true", func(t *testing.T) {
		ok, err := e.Catch(Atom("true"), Atom("b"), Atom("fail"), Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("false", func(t *testing.T) {
		ok, err := e.Catch(Atom("fail"), Atom("b"), Atom("fail"), Done)
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestUnifyWithOccursCheck(t *testing.T) {
	v := Variable{Name: "X"}
	ok, err := UnifyWithOccursCheck(&v, &Compound{
		Functor: "f",
		Args:    []Term{&v},
	}, Done)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestEngine_CurrentPredicate(t *testing.T) {
	e := Engine{procedures: map[string]procedure{
		"(=)/2": nil,
	}}

	var v Variable
	ok, err := e.CurrentPredicate(&v, Done)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, &Compound{
		Functor: "/",
		Args: []Term{
			Atom("="),
			Integer(2),
		},
	}, v.Ref)

	ok, err = e.CurrentPredicate(&v, func() (bool, error) {
		return false, nil
	})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestEngine_Assertz(t *testing.T) {
	var e Engine

	ok, err := e.Assertz(&Compound{
		Functor: "foo",
		Args:    []Term{Atom("a")},
	}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = e.Assertz(&Compound{
		Functor: "foo",
		Args:    []Term{Atom("b")},
	}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	var c int
	ok, err = e.Query("foo(X).", func(vars []*Variable) bool {
		switch c {
		case 0:
			assert.Equal(t, &Variable{Name: "X", Ref: Atom("a")}, vars[0])
		case 1:
			assert.Equal(t, &Variable{Name: "X", Ref: Atom("b")}, vars[0])
		default:
			assert.Fail(t, "unreachable")
		}
		c++
		return false
	})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestEngine_Asserta(t *testing.T) {
	var e Engine

	ok, err := e.Asserta(&Compound{
		Functor: "foo",
		Args:    []Term{Atom("a")},
	}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = e.Asserta(&Compound{
		Functor: "foo",
		Args:    []Term{Atom("b")},
	}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	var c int
	ok, err = e.Query("foo(X).", func(vars []*Variable) bool {
		switch c {
		case 0:
			assert.Equal(t, &Variable{Name: "X", Ref: Atom("b")}, vars[0])
		case 1:
			assert.Equal(t, &Variable{Name: "X", Ref: Atom("a")}, vars[0])
		default:
			assert.Fail(t, "unreachable")
		}
		c++
		return false
	})
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestEngine_Retract(t *testing.T) {
	t.Run("retract the first one", func(t *testing.T) {
		e, err := NewEngine(nil, nil)
		assert.NoError(t, err)
		assert.NoError(t, e.Load("foo(a)."))
		assert.NoError(t, e.Load("foo(b)."))
		assert.NoError(t, e.Load("foo(c)."))
		ok, err := e.Query("retract(foo(X)).", func([]*Variable) bool {
			return true
		})
		assert.NoError(t, err)
		assert.True(t, ok)

		c := 0
		ok, err = e.Query("foo(X).", func(vars []*Variable) bool {
			switch c {
			case 0:
				assert.Equal(t, []*Variable{{Name: "X", Ref: Atom("b")}}, vars)
			case 1:
				assert.Equal(t, []*Variable{{Name: "X", Ref: Atom("c")}}, vars)
			default:
				assert.Fail(t, "unreachable")
			}
			c++
			return false
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("retract the specific one", func(t *testing.T) {
		e, err := NewEngine(nil, nil)
		assert.NoError(t, err)
		assert.NoError(t, e.Load("foo(a)."))
		assert.NoError(t, e.Load("foo(b)."))
		assert.NoError(t, e.Load("foo(c)."))
		ok, err := e.Query("retract(foo(b)).", func([]*Variable) bool {
			return true
		})
		assert.NoError(t, err)
		assert.True(t, ok)

		c := 0
		ok, err = e.Query("foo(X).", func(vars []*Variable) bool {
			switch c {
			case 0:
				assert.Equal(t, []*Variable{{Name: "X", Ref: Atom("a")}}, vars)
			case 1:
				assert.Equal(t, []*Variable{{Name: "X", Ref: Atom("c")}}, vars)
			default:
				assert.Fail(t, "unreachable")
			}
			c++
			return false
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("retract all", func(t *testing.T) {
		e, err := NewEngine(nil, nil)
		assert.NoError(t, err)
		assert.NoError(t, e.Load("foo(a)."))
		assert.NoError(t, e.Load("foo(b)."))
		assert.NoError(t, e.Load("foo(c)."))
		ok, err := e.Query("retract(foo(X)).", func([]*Variable) bool {
			return false
		})
		assert.NoError(t, err)
		assert.False(t, ok)

		ok, err = e.Query("foo(X).", func([]*Variable) bool {
			assert.Fail(t, "unreachable")
			return true
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("variable", func(t *testing.T) {
		e, err := NewEngine(nil, nil)
		assert.NoError(t, err)
		assert.NoError(t, e.Load("foo(a)."))
		assert.NoError(t, e.Load("foo(b)."))
		assert.NoError(t, e.Load("foo(c)."))
		_, err = e.Query("retract(X).", func([]*Variable) bool {
			return false
		})
		assert.Error(t, err)
	})

	t.Run("no clause matches", func(t *testing.T) {
		e, err := NewEngine(nil, nil)
		assert.NoError(t, err)
		ok, err := e.Query("retract(foo(X)).", func([]*Variable) bool {
			return true
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("builtin", func(t *testing.T) {
		e, err := NewEngine(nil, nil)
		assert.NoError(t, err)
		_, err = e.Query("retract(call(X)).", func([]*Variable) bool {
			return true
		})
		assert.Error(t, err)
	})

	t.Run("exception in continuation", func(t *testing.T) {
		e, err := NewEngine(nil, nil)
		assert.NoError(t, err)
		assert.NoError(t, e.Load("foo(a)."))
		_, err = e.Query("retract(foo(X)), throw(e).", func([]*Variable) bool {
			return false
		})
		assert.Error(t, err)

		// removed
		ok, err := e.Query("foo(a).", func([]*Variable) bool {
			assert.Fail(t, "unreachable")
			return true
		})
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestEngine_Abolish(t *testing.T) {
	e, err := NewEngine(nil, nil)
	assert.NoError(t, err)
	assert.NoError(t, e.Load("foo(a)."))
	assert.NoError(t, e.Load("foo(b)."))
	assert.NoError(t, e.Load("foo(c)."))

	ok, err := e.Abolish(&Compound{
		Functor: "/",
		Args:    []Term{Atom("foo"), Integer(1)},
	}, Done)
	assert.NoError(t, err)
	assert.True(t, ok)

	_, ok = e.procedures["foo/1"]
	assert.False(t, ok)
}

func TestEngine_CurrentInput(t *testing.T) {
	var buf bytes.Buffer
	e, err := NewEngine(&buf, nil)
	assert.NoError(t, err)
	_, err = e.Query("current_input(X).", func(vars []*Variable) bool {
		assert.Equal(t, &Variable{
			Name: "X",
			Ref:  &Variable{Ref: Stream{ReadWriteCloser: &input{Reader: &buf}}},
		}, vars[0])
		return true
	})
	assert.NoError(t, err)
}

func TestEngine_CurrentOutput(t *testing.T) {
	var buf bytes.Buffer
	e, err := NewEngine(nil, &buf)
	assert.NoError(t, err)
	_, err = e.Query("current_output(X).", func(vars []*Variable) bool {
		assert.Equal(t, &Variable{
			Name: "X",
			Ref:  &Variable{Ref: Stream{ReadWriteCloser: &output{Writer: &buf}}},
		}, vars[0])
		return true
	})
	assert.NoError(t, err)
}

func TestEngine_SetInput(t *testing.T) {
	t.Run("stream", func(t *testing.T) {
		var e Engine
		s := Stream{ReadWriteCloser: os.Stdin}
		ok, err := e.SetInput(&Variable{Ref: s}, Done)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, s, e.input)
	})

	t.Run("atom defined as a stream global variable", func(t *testing.T) {
		s := Stream{ReadWriteCloser: os.Stdin}
		e := Engine{
			globalVars: map[Atom]Term{
				"x": s,
			},
		}
		ok, err := e.SetInput(&Variable{Ref: Atom("x")}, Done)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, s, e.input)
	})

	t.Run("atom defined as a non-stream global variable", func(t *testing.T) {
		e := Engine{
			globalVars: map[Atom]Term{
				"x": Integer(1),
			},
		}
		_, err := e.SetInput(&Variable{Ref: Atom("x")}, Done)
		assert.Error(t, err)
	})

	t.Run("atom not defined as a global variable", func(t *testing.T) {
		var e Engine
		_, err := e.SetInput(&Variable{Ref: Atom("x")}, Done)
		assert.Error(t, err)
	})
}

func TestEngine_SetOutput(t *testing.T) {
	t.Run("stream", func(t *testing.T) {
		var e Engine
		s := Stream{ReadWriteCloser: os.Stdout}
		ok, err := e.SetOutput(&Variable{Ref: s}, Done)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, s, e.output)
	})

	t.Run("atom defined as a stream global variable", func(t *testing.T) {
		s := Stream{ReadWriteCloser: os.Stdout}
		e := Engine{
			globalVars: map[Atom]Term{
				"x": s,
			},
		}
		ok, err := e.SetOutput(&Variable{Ref: Atom("x")}, Done)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, s, e.output)
	})

	t.Run("atom defined as a non-stream global variable", func(t *testing.T) {
		e := Engine{
			globalVars: map[Atom]Term{
				"x": Integer(1),
			},
		}
		_, err := e.SetOutput(&Variable{Ref: Atom("x")}, Done)
		assert.Error(t, err)
	})

	t.Run("atom not defined as a global variable", func(t *testing.T) {
		var e Engine
		_, err := e.SetOutput(&Variable{Ref: Atom("x")}, Done)
		assert.Error(t, err)
	})
}

func TestEngine_Open(t *testing.T) {
	var e Engine

	t.Run("read", func(t *testing.T) {
		f, err := ioutil.TempFile("", "open_test_read")
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.Remove(f.Name()))
		}()

		_, err = fmt.Fprintf(f, "test\n")
		assert.NoError(t, err)

		assert.NoError(t, f.Close())

		var v Variable
		ok, err := e.Open(Atom(f.Name()), Atom("read"), &v, List(&Compound{
			Functor: "alias",
			Args:    []Term{Atom("input")},
		}), Done)
		assert.NoError(t, err)
		assert.True(t, ok)

		s, ok := v.Ref.(Stream)
		assert.True(t, ok)

		assert.Equal(t, e.globalVars["input"], s)

		b, err := ioutil.ReadAll(s)
		assert.NoError(t, err)
		assert.Equal(t, "test\n", string(b))
	})

	t.Run("write", func(t *testing.T) {
		n := filepath.Join(os.TempDir(), "open_test_write")
		defer func() {
			assert.NoError(t, os.Remove(n))
		}()

		var v Variable
		ok, err := e.Open(Atom(n), Atom("write"), &v, List(&Compound{
			Functor: "alias",
			Args:    []Term{Atom("output")},
		}), Done)
		assert.NoError(t, err)
		assert.True(t, ok)

		s, ok := v.Ref.(Stream)
		assert.True(t, ok)

		assert.Equal(t, e.globalVars["output"], s)

		_, err = fmt.Fprintf(s, "test\n")
		assert.NoError(t, err)

		f, err := os.Open(n)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, f.Close())
		}()

		b, err := ioutil.ReadAll(f)
		assert.NoError(t, err)
		assert.Equal(t, "test\n", string(b))
	})

	t.Run("append", func(t *testing.T) {
		f, err := ioutil.TempFile("", "open_test_append")
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.Remove(f.Name()))
		}()

		_, err = fmt.Fprintf(f, "test\n")
		assert.NoError(t, err)

		assert.NoError(t, f.Close())

		var v Variable
		ok, err := e.Open(Atom(f.Name()), Atom("append"), &v, List(&Compound{
			Functor: "alias",
			Args:    []Term{Atom("append")},
		}), Done)
		assert.NoError(t, err)
		assert.True(t, ok)

		s, ok := v.Ref.(Stream)
		assert.True(t, ok)

		assert.Equal(t, e.globalVars["append"], s)

		_, err = fmt.Fprintf(s, "test\n")
		assert.NoError(t, err)

		f, err = os.Open(f.Name())
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, f.Close())
		}()

		b, err := ioutil.ReadAll(f)
		assert.NoError(t, err)
		assert.Equal(t, "test\ntest\n", string(b))
	})

	t.Run("invalid file name", func(t *testing.T) {
		var v Variable
		_, err := e.Open(&Variable{}, Atom("read"), &v, List(&Compound{
			Functor: "alias",
			Args:    []Term{Atom("input")},
		}), Done)
		assert.Error(t, err)
	})

	t.Run("invalid mode", func(t *testing.T) {
		var v Variable
		_, err := e.Open(Atom("/dev/null"), Atom("invalid"), &v, List(&Compound{
			Functor: "alias",
			Args:    []Term{Atom("input")},
		}), Done)
		assert.Error(t, err)
	})

	t.Run("invalid alias", func(t *testing.T) {
		var v Variable
		_, err := e.Open(Atom("/dev/null"), Atom("read"), &v, List(&Compound{
			Functor: "alias",
			Args:    []Term{&Variable{}},
		}), Done)
		assert.Error(t, err)
	})

	t.Run("unknown option", func(t *testing.T) {
		var v Variable
		_, err := e.Open(Atom("/dev/null"), Atom("read"), &v, List(&Compound{
			Functor: "unknown",
			Args:    []Term{Atom("option")},
		}), Done)
		assert.Error(t, err)
	})
}

func TestEngine_Close(t *testing.T) {
	t.Run("without options", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			var m mockReadWriteCloser
			m.On("Close").Return(nil).Once()
			defer m.AssertExpectations(t)

			var e Engine
			ok, err := e.Close(Stream{ReadWriteCloser: &m}, List(), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})

		t.Run("ng", func(t *testing.T) {
			var m mockReadWriteCloser
			m.On("Close").Return(errors.New("")).Once()
			defer m.AssertExpectations(t)

			var e Engine
			_, err := e.Close(Stream{ReadWriteCloser: &m}, List(), Done)
			assert.Error(t, err)
		})
	})

	t.Run("force false", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			var m mockReadWriteCloser
			m.On("Close").Return(nil).Once()
			defer m.AssertExpectations(t)

			var e Engine
			ok, err := e.Close(Stream{ReadWriteCloser: &m}, List(&Compound{
				Functor: "force",
				Args:    []Term{Atom("false")},
			}), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})

		t.Run("ng", func(t *testing.T) {
			var m mockReadWriteCloser
			m.On("Close").Return(errors.New("")).Once()
			defer m.AssertExpectations(t)

			var e Engine
			_, err := e.Close(Stream{ReadWriteCloser: &m}, List(&Compound{
				Functor: "force",
				Args:    []Term{Atom("false")},
			}), Done)
			assert.Error(t, err)
		})
	})

	t.Run("force true", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			var m mockReadWriteCloser
			m.On("Close").Return(nil).Once()
			defer m.AssertExpectations(t)

			var e Engine
			ok, err := e.Close(Stream{ReadWriteCloser: &m}, List(&Compound{
				Functor: "force",
				Args:    []Term{Atom("true")},
			}), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})

		t.Run("ng", func(t *testing.T) {
			var m mockReadWriteCloser
			m.On("Close").Return(errors.New("")).Once()
			defer m.AssertExpectations(t)

			var e Engine
			ok, err := e.Close(Stream{ReadWriteCloser: &m}, List(&Compound{
				Functor: "force",
				Args:    []Term{Atom("true")},
			}), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})

	t.Run("valid global variable", func(t *testing.T) {
		var m mockReadWriteCloser
		m.On("Close").Return(nil).Once()
		defer m.AssertExpectations(t)

		e := Engine{
			globalVars: map[Atom]Term{
				"foo": Stream{ReadWriteCloser: &m},
			},
		}
		ok, err := e.Close(Atom("foo"), List(), Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unknown global variable", func(t *testing.T) {
		var e Engine
		_, err := e.Close(Atom("foo"), List(), Done)
		assert.Error(t, err)
	})

	t.Run("non stream", func(t *testing.T) {
		var e Engine
		_, err := e.Close(&Variable{}, List(), Done)
		assert.Error(t, err)
	})

	t.Run("unknown option", func(t *testing.T) {
		var e Engine
		_, err := e.Close(Stream{}, List(&Compound{
			Functor: "unknown",
			Args:    []Term{Atom("option")},
		}), Done)
		assert.Error(t, err)
	})
}

type mockReadWriteCloser struct {
	mock.Mock
}

func (m *mockReadWriteCloser) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockReadWriteCloser) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockReadWriteCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestEngine_FlushOutput(t *testing.T) {
	t.Run("non flusher", func(t *testing.T) {
		var m mockReadWriteCloser
		defer m.AssertExpectations(t)

		var e Engine
		ok, err := e.FlushOutput(Stream{ReadWriteCloser: &m}, Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("flusher", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			var m struct {
				mockReadWriteCloser
				mockFlusher
			}
			m.mockFlusher.On("Flush").Return(nil).Once()
			defer m.mockReadWriteCloser.AssertExpectations(t)
			defer m.mockFlusher.AssertExpectations(t)

			var e Engine
			ok, err := e.FlushOutput(Stream{ReadWriteCloser: &m}, Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})

		t.Run("ng", func(t *testing.T) {
			var m struct {
				mockReadWriteCloser
				mockFlusher
			}
			m.mockFlusher.On("Flush").Return(errors.New("")).Once()
			defer m.mockReadWriteCloser.AssertExpectations(t)
			defer m.mockFlusher.AssertExpectations(t)

			var e Engine
			_, err := e.FlushOutput(Stream{ReadWriteCloser: &m}, Done)
			assert.Error(t, err)
		})
	})

	t.Run("valid global variable", func(t *testing.T) {
		var m mockReadWriteCloser
		defer m.AssertExpectations(t)

		e := Engine{
			globalVars: map[Atom]Term{
				"foo": Stream{ReadWriteCloser: &m},
			},
		}
		ok, err := e.FlushOutput(Atom("foo"), Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unknown global variable", func(t *testing.T) {
		var e Engine
		_, err := e.FlushOutput(Atom("foo"), Done)
		assert.Error(t, err)
	})

	t.Run("non stream", func(t *testing.T) {
		var e Engine
		_, err := e.FlushOutput(&Variable{}, Done)
		assert.Error(t, err)
	})
}

type mockFlusher struct {
	mock.Mock
}

func (m *mockFlusher) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func TestEngine_WriteTerm(t *testing.T) {
	var io mockReadWriteCloser
	defer io.AssertExpectations(t)

	s := Stream{ReadWriteCloser: &io}

	ops := operators{
		{Precedence: 500, Type: "yfx", Name: "+"},
		{Precedence: 200, Type: "fy", Name: "-"},
	}

	e := Engine{
		operators: ops,
		globalVars: map[Atom]Term{
			"foo": s,
		},
	}

	t.Run("without options", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			var m mockTerm
			m.On("WriteTerm", s, WriteTermOptions{ops: ops}).Return(nil).Once()
			defer m.AssertExpectations(t)

			ok, err := e.WriteTerm(s, &m, List(), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})

		t.Run("ng", func(t *testing.T) {
			var m mockTerm
			m.On("WriteTerm", s, WriteTermOptions{ops: ops}).Return(errors.New("")).Once()
			defer m.AssertExpectations(t)

			_, err := e.WriteTerm(s, &m, List(), Done)
			assert.Error(t, err)
		})
	})

	t.Run("quoted", func(t *testing.T) {
		t.Run("false", func(t *testing.T) {
			var m mockTerm
			m.On("WriteTerm", s, WriteTermOptions{quoted: false, ops: ops}).Return(nil).Once()
			defer m.AssertExpectations(t)

			ok, err := e.WriteTerm(s, &m, List(&Compound{
				Functor: "quoted",
				Args:    []Term{Atom("false")},
			}), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})

		t.Run("true", func(t *testing.T) {
			var m mockTerm
			m.On("WriteTerm", s, WriteTermOptions{quoted: true, ops: ops}).Return(nil).Once()
			defer m.AssertExpectations(t)

			ok, err := e.WriteTerm(s, &m, List(&Compound{
				Functor: "quoted",
				Args:    []Term{Atom("true")},
			}), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})

	t.Run("ignore_ops", func(t *testing.T) {
		t.Run("false", func(t *testing.T) {
			var m mockTerm
			m.On("WriteTerm", s, WriteTermOptions{ops: ops}).Return(nil).Once()
			defer m.AssertExpectations(t)

			ok, err := e.WriteTerm(s, &m, List(&Compound{
				Functor: "ignore_ops",
				Args:    []Term{Atom("false")},
			}), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})

		t.Run("true", func(t *testing.T) {
			var m mockTerm
			m.On("WriteTerm", s, WriteTermOptions{ops: nil}).Return(nil).Once()
			defer m.AssertExpectations(t)

			ok, err := e.WriteTerm(s, &m, List(&Compound{
				Functor: "ignore_ops",
				Args:    []Term{Atom("true")},
			}), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})

	t.Run("numbervars", func(t *testing.T) {
		t.Run("false", func(t *testing.T) {
			var m mockTerm
			m.On("WriteTerm", s, WriteTermOptions{ops: ops, numberVars: false}).Return(nil).Once()
			defer m.AssertExpectations(t)

			ok, err := e.WriteTerm(s, &m, List(&Compound{
				Functor: "numbervars",
				Args:    []Term{Atom("false")},
			}), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})

		t.Run("true", func(t *testing.T) {
			var m mockTerm
			m.On("WriteTerm", s, WriteTermOptions{ops: ops, numberVars: true}).Return(nil).Once()
			defer m.AssertExpectations(t)

			ok, err := e.WriteTerm(s, &m, List(&Compound{
				Functor: "numbervars",
				Args:    []Term{Atom("true")},
			}), Done)
			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})

	t.Run("unknown option", func(t *testing.T) {
		var m mockTerm
		defer m.AssertExpectations(t)

		_, err := e.WriteTerm(s, &m, List(&Compound{
			Functor: "unknown",
			Args:    []Term{Atom("option")},
		}), Done)
		assert.Error(t, err)
	})

	t.Run("valid global variable", func(t *testing.T) {
		var m mockTerm
		m.On("WriteTerm", s, WriteTermOptions{ops: ops}).Return(nil).Once()
		defer m.AssertExpectations(t)

		ok, err := e.WriteTerm(Atom("foo"), &m, List(), Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unknown global variable", func(t *testing.T) {
		var m mockTerm
		defer m.AssertExpectations(t)

		_, err := e.WriteTerm(Atom("bar"), &m, List(), Done)
		assert.Error(t, err)
	})

	t.Run("non stream", func(t *testing.T) {
		var m mockTerm
		defer m.AssertExpectations(t)

		_, err := e.WriteTerm(&Variable{}, &m, List(), Done)
		assert.Error(t, err)
	})
}

type mockTerm struct {
	mock.Mock
}

func (m *mockTerm) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockTerm) WriteTerm(w io.Writer, opts WriteTermOptions) error {
	args := m.Called(w, opts)
	return args.Error(0)
}

func (m *mockTerm) Unify(t Term, occursCheck bool) bool {
	args := m.Called(t, occursCheck)
	return args.Bool(0)
}

func (m *mockTerm) Copy() Term {
	args := m.Called()
	return args.Get(0).(Term)
}

func TestCharCode(t *testing.T) {
	t.Run("ascii", func(t *testing.T) {
		ok, err := CharCode(Atom("a"), Integer(97), Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("emoji", func(t *testing.T) {
		ok, err := CharCode(Atom("😀"), Integer(128512), Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("query char", func(t *testing.T) {
		var v Variable
		ok, err := CharCode(&v, Integer(128512), Done)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, Atom("😀"), v.Ref)
	})

	t.Run("query code", func(t *testing.T) {
		var v Variable
		ok, err := CharCode(Atom("😀"), &v, Done)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, Integer(128512), v.Ref)
	})

	t.Run("not a character", func(t *testing.T) {
		var v Variable
		_, err := CharCode(Atom("abc"), &v, Done)
		assert.Error(t, err)
	})

	t.Run("not a code", func(t *testing.T) {
		var v Variable
		_, err := CharCode(&v, Float(1.0), Done)
		assert.Error(t, err)
	})
}

func TestEngine_PutByte(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var io mockReadWriteCloser
		io.On("Write", []byte{97}).Return(1, nil).Once()
		defer io.AssertExpectations(t)

		s := Stream{ReadWriteCloser: &io}

		var e Engine
		ok, err := e.PutByte(s, Integer(97), Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("ng", func(t *testing.T) {
		var io mockReadWriteCloser
		io.On("Write", []byte{97}).Return(0, errors.New("")).Once()
		defer io.AssertExpectations(t)

		s := Stream{ReadWriteCloser: &io}

		var e Engine
		_, err := e.PutByte(s, Integer(97), Done)
		assert.Error(t, err)
	})

	t.Run("valid global variable", func(t *testing.T) {
		var io mockReadWriteCloser
		io.On("Write", []byte{97}).Return(1, nil).Once()
		defer io.AssertExpectations(t)

		s := Stream{ReadWriteCloser: &io}

		e := Engine{
			globalVars: map[Atom]Term{
				"foo": s,
			},
		}
		ok, err := e.PutByte(Atom("foo"), Integer(97), Done)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unknown global variable", func(t *testing.T) {
		var e Engine
		_, err := e.PutByte(Atom("foo"), Integer(97), Done)
		assert.Error(t, err)
	})

	t.Run("not a stream", func(t *testing.T) {
		var e Engine
		_, err := e.PutByte(&Variable{}, Integer(97), Done)
		assert.Error(t, err)
	})

	t.Run("not a byte", func(t *testing.T) {
		var io mockReadWriteCloser
		defer io.AssertExpectations(t)

		s := Stream{ReadWriteCloser: &io}

		t.Run("not an integer", func(t *testing.T) {
			var e Engine
			_, err := e.PutByte(s, Atom("a"), Done)
			assert.Error(t, err)
		})

		t.Run("negative", func(t *testing.T) {
			var e Engine
			_, err := e.PutByte(s, Integer(-1), Done)
			assert.Error(t, err)
		})

		t.Run("more than 255", func(t *testing.T) {
			var e Engine
			_, err := e.PutByte(s, Integer(256), Done)
			assert.Error(t, err)
		})
	})
}
