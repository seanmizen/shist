package model

type Entry struct {
	Index     int
	Timestamp int64
	Elapsed   int64
	Command   string
	Raw       string
}