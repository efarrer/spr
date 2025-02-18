package concurrent

func SliceMap[I any, O any](in []I, fn func(I) (O, error)) ([]O, error) {
	type res struct {
		val O
		err error
	}
	ch := make(chan res)
	defer func() { close(ch) }()

	for _, i := range in {
		go func() {
			o, err := fn(i)
			ch <- res{val: o, err: err}
		}()
	}

	var err error
	out := make([]O, 0, len(in))
	for _ = range in {
		res := <-ch
		if res.err != nil {
			err = res.err
		}
		out = append(out, res.val)
	}

	return out, err
}
