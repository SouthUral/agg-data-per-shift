package psql

type listenAndServeError struct {
}

func (e listenAndServeError) Error() string {
	return "error in listenAndServe"
}
