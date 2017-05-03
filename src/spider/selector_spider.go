package spider

type DomSelector struct {
	selector  string
	attribute map[string]string
}

type SelectorSpider struct {
	url      string
	selector DomSelector
	filter   Filter
}

func (ss *SelectorSpider) Spide() chan string {
	out := make(chan string)
	return out
}

func (ss *SelectorSpider) RegisterFilter(filter Filter) {
	ss.filter = filter
}
