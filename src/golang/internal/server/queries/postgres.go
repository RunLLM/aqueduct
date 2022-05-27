package queries

type postgresReaderImpl struct {
	standardReaderImpl
}

func newPostgresReader() Reader {
	return &postgresReaderImpl{standardReaderImpl{}}
}
