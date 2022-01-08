package lists

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/sorting"
)

type Item struct {
	gorm.Model
	sorting.SortingDESC

	Name    string
	Checked bool
	ListID  uint
}
