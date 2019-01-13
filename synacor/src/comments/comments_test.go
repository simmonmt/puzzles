package comments

import (
	"reflect"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	in := `
		3 // foo
		// bar
		// baz
		5-15 
		// two
		10 // ten
	`

	cMap, err := Read(strings.NewReader(in))
	if err != nil {
		t.Errorf("read failed: %v", err)
		return
	}

	expected := map[int][]*Comment{
		3: []*Comment{
			&Comment{Single, 3, 3, []string{"foo"}},
			&Comment{Block, 3, 3, []string{"bar", "baz"}},
		},
		5: []*Comment{
			&Comment{Block, 5, 15, []string{"two"}},
		},
		10: []*Comment{
			&Comment{Single, 10, 10, []string{"ten"}},
		},
	}

	if !reflect.DeepEqual(cMap, expected) {
		t.Errorf("Read() = %+v, want %+v", cMap, expected)
	}
}
