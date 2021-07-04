package network

type Network interface {
    Serve(addr ...string) error
}