package provider

import (
	"strings"
	"testing"
)

func TestSuppressDiffForMultivalueAttributeOrder(t *testing.T) {
	t.Parallel()

	suppress := suppressDiffForMultivalueAttributeOrder()

	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		{
			name:     "identical single value",
			old:      "role-1",
			new:      "role-1",
			expected: true,
		},
		{
			name:     "different single value",
			old:      "role-1",
			new:      "role-2",
			expected: false,
		},
		{
			name:     "identical order",
			old:      joinAttributeValues("role-1", "role-2", "role-3"),
			new:      joinAttributeValues("role-1", "role-2", "role-3"),
			expected: true,
		},
		{
			name:     "same values in a different order",
			old:      joinAttributeValues("role-3", "role-1", "role-2"),
			new:      joinAttributeValues("role-1", "role-2", "role-3"),
			expected: true,
		},
		{
			name:     "value added",
			old:      joinAttributeValues("role-1", "role-2"),
			new:      joinAttributeValues("role-1", "role-2", "role-3"),
			expected: false,
		},
		{
			name:     "value removed",
			old:      joinAttributeValues("role-1", "role-2", "role-3"),
			new:      joinAttributeValues("role-1", "role-3"),
			expected: false,
		},
		{
			name:     "value replaced, count unchanged",
			old:      joinAttributeValues("role-1", "role-2"),
			new:      joinAttributeValues("role-1", "role-3"),
			expected: false,
		},
		{
			name:     "duplicates are significant",
			old:      joinAttributeValues("role-1", "role-1"),
			new:      joinAttributeValues("role-1", "role-2"),
			expected: false,
		},
		{
			name:     "reordered duplicates",
			old:      joinAttributeValues("role-2", "role-1", "role-1"),
			new:      joinAttributeValues("role-1", "role-1", "role-2"),
			expected: true,
		},
		{
			name:     "attribute created",
			old:      "",
			new:      joinAttributeValues("role-1", "role-2"),
			expected: false,
		},
		{
			name:     "attribute cleared",
			old:      joinAttributeValues("role-1", "role-2"),
			new:      "",
			expected: false,
		},
		{
			name:     "both empty",
			old:      "",
			new:      "",
			expected: true,
		},
		// the DiffSuppressFunc of a TypeMap is also handed the flatmap element count
		// key ("attributes.%"), which must keep its diff whenever the count changes
		{
			name:     "element count unchanged",
			old:      "3",
			new:      "3",
			expected: true,
		},
		{
			name:     "element count changed",
			old:      "2",
			new:      "3",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if actual := suppress("attributes.some_attribute", test.old, test.new, nil); actual != test.expected {
				t.Errorf("expected suppression of %q -> %q to be %t, but was %t", test.old, test.new, test.expected, actual)
			}
		})
	}
}

func joinAttributeValues(values ...string) string {
	return strings.Join(values, MULTIVALUE_ATTRIBUTE_SEPARATOR)
}
