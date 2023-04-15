package menu

type NavigationType int64

const (
	Stop     NavigationType = 0
	Continue NavigationType = 1
	Replay   NavigationType = 2
)

type Navigation struct {
	selections []string
}

func NewNavigation() *Navigation {
	return &Navigation{}
}

func (n *Navigation) GoBack() {
	i := len(n.selections) - 1
	n.selections[i] = n.selections[len(n.selections)-1] // Copy last element to index i.
	n.selections[len(n.selections)-1] = ""              // Erase last element (write zero value).
	n.selections = n.selections[:len(n.selections)-1]
}

func (n *Navigation) AddToSelections(s string) {
	n.selections = append(n.selections, s)
}

func (n *Navigation) GetSelections() []string {
	return n.selections
}
