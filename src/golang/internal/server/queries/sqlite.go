package queries

type sqliteReaderImpl struct {
	standardReaderImpl
}

func newSqliteReader() Reader {
	return &sqliteReaderImpl{standardReaderImpl{}}
}
