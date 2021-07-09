package net

type Network interface {
    Serve(addr ...string) error
}