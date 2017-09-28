package container

type RootFSProvider interface {
	PullRootFS() error
}
