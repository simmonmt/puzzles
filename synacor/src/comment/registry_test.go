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

	cs, err := Read(strings.NewReader(in))
	if err != nil {
		t.Errorf("read failed: %v", err)
		return
	}

	expected := Registry{
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

	if !reflect.DeepEqual(cs, expected) {
		t.Errorf("Read() = %+v, want %+v", cs, expected)
		return
	}

	if comment, found := cs.GetSingle(3); !found || comment != "foo" {
		t.Errorf(`GetSingle(3) = "%s", %v, want "foo", true`, comment, found)
	}
	if comment, found := cs.GetSingle(4); found {
		t.Errorf(`GetSingle(3) = "%s", %v, want _, false`, comment, found)
	}

	expectedBlock := &Comment{Block, 3, 3, []string{"bar", "baz"}}
	if block := cs.GetBlock(3); !reflect.DeepEqual(block, expectedBlock) {
		t.Errorf(`GetBlock(3) = %+v, want %+v`, block, expectedBlock)
	}
	if block := cs.GetBlock(6); block != nil {
		t.Errorf(`GetBlock(6) = %+v, want nil`, block)
	}
}
