package errorz

import (
	"errors"
	"fmt"
	"testing"
)

var errT = errors.New("var error type")

func TestErrShallow(t *testing.T) {
	e := a()
	et := e.(ErrMadNet)
	fmt.Printf("%s\n", et)
}

func TestErrDeep(t *testing.T) {
	e := aa()
	et := e.(ErrMadNet)
	fmt.Printf("%s\n", et)
}

func TestErrShallowNew(t *testing.T) {
	e := aaa()
	et := e.(ErrMadNet)
	fmt.Printf("%s\n", et)
}

func TestErrNoTrace(t *testing.T) {
	e := Wrap(func() error {
		return Wrap(func() error {
			return Wrap(func() error {
				return Wrap(func() error {
					return New("Turtles").WithContext(" all the way down.")
				}()).WithContext("%d", 1)
			}()).WithContext("%d", 2)
		}()).WithContext("%d", 3)
	}()).WithContext("%d", 4)
	et := e.(ErrMadNet)
	fmt.Printf("%s\n", et)
}

func TestInlineFN(t *testing.T) {
	e := Wrap(func() error {
		return Wrap(func() error {
			return Wrap(func() error {
				return Wrap(func() error {
					return New("Turtles").WithContext(" all the way down.").WithTrace()
				}()).WithContext("%d", 1).WithTrace()
			}()).WithContext("%d", 2).WithTrace()
		}()).WithContext("%d", 3).WithTrace()
	}()).WithContext("%d", 4).WithTrace()
	et := e.(ErrMadNet)
	fmt.Printf("%s\n", et)
}

func TestIterFn(t *testing.T) {
	err := Newf("%v", "Turtles").WithContext(" all the way down.").WithTrace()
	for i := 0; i < 1000; i++ {
		err = Wrap(err).WithContext(" %d ", i)
	}
	et := err.(ErrMadNet)
	fmt.Printf("%s\n", et)
}

func TestAsIs(t *testing.T) {
	eRR := errors.New("foobar")
	ee := Wrap(eRR)
	eee := Wrap(ee)
	eeee := Wrap(eee)
	e := Wrap(eeee)
	// non-wrapped pointer with as should fail
	if As(eRR, ErrMadNetType()) {
		t.Fatal("")
	}
	// pointer with as should not fail for type
	if !As(e, ErrMadNetType()) {
		t.Fatal("")
	}
	// pointer with as should not fail for wrapped type
	if !As(e, &eRR) {
		t.Fatal("")
	}
	// pointers with is should fail
	if Is(eRR, *ErrMadNetType()) {
		t.Fatal("")
	}
	// pointers with is should fail
	if Is(e, *ErrMadNetType()) {
		t.Fatal("")
	}
	// vars with is should not fail for correct var
	if !Is(e, eRR) {
		t.Fatal("")
	}
}

func a() error {
	if err := b(); err != nil {
		return Wrap(err).WithTrace()
	}
	return nil
}

func b() error {
	if err := c(); err != nil {
		return Wrap(err)
	}
	return nil
}

func c() error {
	if err := d(); err != nil {
		return Wrap(err).WithTrace()
	}
	return nil
}

func d() error {
	return errT
}

func aa() error {
	if err := bb(); err != nil {
		return Wrap(err).WithTrace()
	}
	return nil
}

func bb() error {
	if err := cc(); err != nil {
		return Wrap(err)
	}
	return nil
}

func cc() error {
	if err := dd(); err != nil {
		return Wrap(err).WithTrace()
	}
	return nil
}

func dd() error {
	if err := ee(); err != nil {
		return Wrap(err)
	}
	return nil
}

func ee() error {
	if err := ff(); err != nil {
		return Wrap(err)
	}
	return nil
}

func ff() error {
	if err := gg(); err != nil {
		return Wrap(err).WithTrace()
	}
	return nil
}

func gg() error {
	if err := hh(); err != nil {
		return Wrap(err)
	}
	return nil
}

func hh() error {
	if err := ii(); err != nil {
		return Wrap(err)
	}
	return nil
}

func ii() error {
	if err := jj(); err != nil {
		return Wrap(err).WithTrace()
	}
	return nil
}

func jj() error {
	return errT
}

func aaa() error {
	if err := bbb(); err != nil {
		return Wrap(err)
	}
	return nil
}

func bbb() error {
	if err := ccc(); err != nil {
		return Wrap(err)
	}
	return nil
}

func ccc() error {
	if err := ddd(); err != nil {
		return Wrap(err).WithTrace()
	}
	return nil
}

func ddd() error {
	return New("New error type")
}
