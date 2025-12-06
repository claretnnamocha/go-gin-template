package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go-gin-template/pkg/utils"
)

func TestHashPassword(t *testing.T) {
	// Arrange
	password := "TestPassword123!"

	// Act
	hash, err := utils.HashPassword(password)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestCheckPassword(t *testing.T) {
	// Arrange
	password := "TestPassword123!"
	hash, _ := utils.HashPassword(password)

	// Act & Assert
	assert.True(t, utils.CheckPassword(password, hash))
	assert.False(t, utils.CheckPassword("WrongPassword", hash))
}

func TestGenerateRandomString(t *testing.T) {
	testCases := []struct {
		name   string
		length int
	}{
		{"length 10", 10},
		{"length 20", 20},
		{"length 32", 32},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			str, err := utils.GenerateRandomString(tc.length)

			assert.NoError(t, err)
			assert.Len(t, str, tc.length)
		})
	}
}

func TestSlugify(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"This is a TEST", "this-is-a-test"},
		{"Special!@#$Characters", "specialcharacters"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"  Leading and Trailing  ", "leading-and-trailing"},
		{"already-slugified", "already-slugified"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := utils.Slugify(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	testCases := []struct {
		input    string
		length   int
		expected string
	}{
		{"Hello World", 5, "Hello"},
		{"Short", 10, "Short"},
		{"Exactly Ten", 11, "Exactly Ten"},
		{"", 5, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := utils.TruncateString(tc.input, tc.length)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	testCases := []struct {
		input    string
		length   int
		expected string
	}{
		{"Hello World", 8, "Hello..."},
		{"Short", 10, "Short"},
		{"Hi", 2, "Hi"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := utils.TruncateWithEllipsis(tc.input, tc.length)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsEmpty(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"\t\n", true},
		{"hello", false},
		{" hello ", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := utils.IsEmpty(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	assert.True(t, utils.ContainsString(slice, "banana"))
	assert.False(t, utils.ContainsString(slice, "grape"))
	assert.False(t, utils.ContainsString([]string{}, "apple"))
}

func TestUniqueStrings(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "with duplicates",
			input:    []string{"apple", "banana", "apple", "cherry", "banana"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "no duplicates",
			input:    []string{"apple", "banana", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := utils.UniqueStrings(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMaskEmail(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"test@example.com", "t**t@example.com"},
		{"ab@example.com", "ab@example.com"},
		{"a@example.com", "a@example.com"},
		{"invalid-email", "invalid-email"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := utils.MaskEmail(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMaskPhone(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"+1234567890", "*******7890"},
		{"12345", "*2345"},
		{"123", "123"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := utils.MaskPhone(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPointerHelpers(t *testing.T) {
	// String pointer
	strPtr := utils.StringPtr("hello")
	assert.NotNil(t, strPtr)
	assert.Equal(t, "hello", *strPtr)

	// Int pointer
	intPtr := utils.IntPtr(42)
	assert.NotNil(t, intPtr)
	assert.Equal(t, 42, *intPtr)

	// Bool pointer
	boolPtr := utils.BoolPtr(true)
	assert.NotNil(t, boolPtr)
	assert.True(t, *boolPtr)
}

func TestDerefHelpers(t *testing.T) {
	// Deref string
	str := "hello"
	assert.Equal(t, "hello", utils.DerefString(&str, "default"))
	assert.Equal(t, "default", utils.DerefString(nil, "default"))

	// Deref int
	num := 42
	assert.Equal(t, 42, utils.DerefInt(&num, 0))
	assert.Equal(t, 0, utils.DerefInt(nil, 0))

	// Deref bool
	b := true
	assert.True(t, utils.DerefBool(&b, false))
	assert.False(t, utils.DerefBool(nil, false))
}

// Benchmark examples
func BenchmarkHashPassword(b *testing.B) {
	password := "TestPassword123!"
	for i := 0; i < b.N; i++ {
		_, _ = utils.HashPassword(password)
	}
}

func BenchmarkSlugify(b *testing.B) {
	input := "This is a Test String for Slugification!"
	for i := 0; i < b.N; i++ {
		_ = utils.Slugify(input)
	}
}
