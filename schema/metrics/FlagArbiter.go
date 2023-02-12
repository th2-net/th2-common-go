package metrics

type set map[interface{}]interface{}

func (st set) contains(item interface{}) bool {
	if _, ok := st[item]; ok {
		return true
	}
	return false
}

func (st set) add(item interface{}) {
	st[item] = struct{}{}
}

func (st set) remove(item interface{}) {
	delete(st, item)
}

func (st set) isEmpty() bool {
	return len(st) == 0
}

type FlagArbiter struct {
	flags []Flag

	disabled set
}

func NewFlagArbiter(flags ...Flag) *FlagArbiter {
	return &FlagArbiter{
		flags:    flags,
		disabled: make(set),
	}
}

func (flagArb *FlagArbiter) RegisterMonitor(name string) *Monitor {
	flagArb.disabled.add(name)
	return &Monitor{
		Name:        name,
		FlagArbiter: flagArb,
	}
}

func (flagArb *FlagArbiter) EnableMonitor(name string) {
	flagArb.disabled.remove(name)
	if flagArb.disabled.isEmpty() {
		flagArb.enableFlags()
	}
}

func (flagArb *FlagArbiter) DisableMonitor(name string) {
	flagArb.disabled.add(name)
	if !flagArb.disabled.isEmpty() {
		flagArb.disableFlags()
	}
}

func (flagArb *FlagArbiter) enableFlags() {
	for _, flag := range flagArb.flags {
		if !flag.IsEnabled() {
			flag.Enable()
		}
	}
}

func (flagArb *FlagArbiter) disableFlags() {
	for _, flag := range flagArb.flags {
		if flag.IsEnabled() {
			flag.Disable()
		}
	}
}
