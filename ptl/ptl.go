package ptl

type Protocol interface {
    Serve(addr ...string) error
    Command(args []string) (interface{}, error)
}