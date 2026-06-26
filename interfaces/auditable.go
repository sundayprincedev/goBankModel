package interfaces

type Auditable interface {
	Audit() []string
}
