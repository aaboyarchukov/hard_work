package mixins

type Reader interface {
	Read(filePath string) []byte
}

type Writer interface {
	Write(content []byte, filePath string) error
}

type User struct {
	Reader
	Writer

	Name string
	Age  int
}

user := User{
	Name: "Alice",
	Age:  11,
}

content := user.Raad(filePath1)
if err := user.Write(content, filePath2); err != nil {
	// ...
}