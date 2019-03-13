package main

import "testing"

var tagsListToStringTests = []struct {
	input  []string
	output string
}{
	{
		[]string{
			"carrot",
			"apple",
			"banana",
		},
		"apple+banana+carrot",
	},
	{
		[]string{
			"potato",
			"apple",
			"tomato",
		},
		"apple+potato+tomato",
	},
}

func TestTagsListToString(t *testing.T) {
	for _, test := range tagsListToStringTests {
		output := tagsListToString(test.input)
		if output != test.output {
			t.Errorf("Output was incorrect, got: %v, want: %v.", output, test.output)
		}
	}
}

var doesMatchTagsTests = []struct {
	searchTags []string
	post       Post
	result     bool
}{
	{
		[]string{"potato"},
		Post{
			Tags: []string{"tomato"},
		},
		false,
	},
	{
		[]string{"-potato"},
		Post{
			Tags: []string{"potato"},
		},
		false,
	},
	{
		[]string{"potato"},
		Post{
			Tags: []string{"potato"},
		},
		true,
	},
	{
		[]string{"*"},
		Post{
			Tags: []string{"potato"},
		},
		true,
	},
	{
		[]string{"-*"},
		Post{
			Tags: []string{"potato"},
		},
		false,
	},
	{
		[]string{""},
		Post{
			Tags: []string{"potato"},
		},
		false,
	},
	{
		[]string{"tomato"},
		Post{
			Tags: []string{"potato", "tomato"},
		},
		true,
	},
	{
		[]string{"tomato"},
		Post{
			Tags: []string{"tomato", "potato"},
		},
		true,
	},
}

func TestDoesMatchTags(t *testing.T) {
	for i, test := range doesMatchTagsTests {
		result := doesMatchTags(test.searchTags, test.post)
		if result != test.result {
			t.Errorf("Test %d, Output was incorrect, got: %t, want: %t.", i, result, test.result)
		}
	}
}

var checkPasswordTests = []struct {
	password1 string
	password2 string
	result    bool
}{
	{
		"$2a$12$IzriL1jyfMJBKoUTJ3i/7eUXQPK/UzzNO6VYev7JmRKvx5fqUIH52",
		"potato",
		true,
	},
	{
		"$2a$12$IzriL1jyfMJBKoUTJ3i/7eUXQPK/UzzNO6VYev7JmRKvx5fqUIH52",
		"",
		false,
	},
	{
		"",
		"",
		false,
	},
	{
		"potato",
		"",
		false,
	},
	{
		"potato",
		"$2a$12$IzriL1jyfMJBKoUTJ3i/7eUXQPK/UzzNO6VYev7JmRKvx5fqUIH52",
		false,
	},
}

func TestCheckPassword(t *testing.T) {
	for i, test := range checkPasswordTests {
		result := checkPassword(test.password1, test.password2)
		if result != test.result {
			t.Errorf("Test %d, Output was incorrect, got: %v, want: %v.", i, result, test.result)
		}
	}
}
