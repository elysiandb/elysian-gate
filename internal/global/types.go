package global

type Node struct {
	Name  string
	Role  string
	HTTP  Transport
	TCP   Transport
	Ready bool
}

type Transport struct {
	Host string
	Port int
	Up   bool
}
