package ports

type NetworkCommands interface {
	Exist(name string) (exist bool, err error)
	Up(name string) (err error)
}
