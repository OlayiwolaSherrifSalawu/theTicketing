package api

type ConstantError string

func (e ConstantError) Error() string {
	return string(e)
}

const (
	OPENING_POOL_FAILED = ConstantError("Failure in creating database pool")
)
