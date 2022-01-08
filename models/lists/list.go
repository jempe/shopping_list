package lists

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/sorting"
)

type List struct {
	gorm.Model
	sorting.SortingDESC

	Name  string
	Items []Item
}
