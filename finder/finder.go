package finder

type Finder interface {
	Find() (error, map[string][]FileInfo)
}
