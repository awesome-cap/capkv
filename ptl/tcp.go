package ptl

type tcp struct {

}

func (t *tcp) Serve(addr ...string) error{
    return nil
}

func (t *tcp) Command(args []string) (interface{}, error){
    return nil, nil
}