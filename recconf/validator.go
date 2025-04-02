package recconf

type Validator interface {
	Validate() error
}
