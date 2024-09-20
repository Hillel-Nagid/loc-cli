package command

import (
	"flag"
	"log"
)

type Command[T any] struct {
	flags   *flag.FlagSet
	target  T
	args    []any
	execute func(T, ...any) error
}

func (c Command[T]) Run() {
	if err := c.execute(c.target, c.args...); err != nil {
		log.Fatal(err)
	}
}
func NewCommand[T any](flags *flag.FlagSet, action func(T, ...any) error, args []any, target T) Command[T] {
	return Command[T]{
		flags:   flags,
		target:  target,
		args:    args,
		execute: action,
	}
}
