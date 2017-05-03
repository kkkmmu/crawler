package spider

type Filter func(match string) bool

type LinkGenerator func(page string, document string) ([]string, error)

type Spider interface {
	Spide(page string)
	RegisterFilter(filter Filter)                  //@liwei: Should be change to setfilter
	RegisterLinkGenerator(generator LinkGenerator) //@liwei: Should be change to setfilter
	Filter(chan string) chan string
	Start() chan string
}
