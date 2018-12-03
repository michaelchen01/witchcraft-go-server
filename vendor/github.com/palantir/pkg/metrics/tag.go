// Copyright (c) 2018 Palantir Technologies. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"strings"

	"github.com/pkg/errors"
)

// A Tag is metadata associated with a metric. This tags implementation is designed to be compatible with the best
// practices for DataDog tags (https://docs.datadoghq.com/guides/tagging/). The key and value for a tag must both be
// non-empty.
type Tag struct {
	key   string
	value string
}

func (t Tag) Key() string {
	return t.key
}

func (t Tag) Value() string {
	return t.value
}

// The full representation of the tag, which is "key:value".
func (t Tag) String() string {
	return t.key + ":" + t.value
}

type Tags []Tag

func (t Tags) ToSet() map[Tag]struct{} {
	tags := make(map[Tag]struct{})
	for _, currTag := range t {
		tags[currTag] = struct{}{}
	}
	return tags
}

// ToMap returns the map representation of the tags, where the map key is the tag key and the map value is the tag
// value. If Tags contains multiple tags with the same key but different values, the output map will only contain one
// entry for the key (and the value will be the last value for that key that appeared in the Tags slice).
func (t Tags) ToMap() map[string]string {
	tags := make(map[string]string)
	for _, currTag := range t {
		tags[currTag.key] = currTag.value
	}
	return tags
}

// MustNewTag returns the result of calling NewTag, but panics if NewTag returns an error. Should only be used in
// instances where the inputs are statically defined and known to be valid.
func MustNewTag(k, v string) Tag {
	t, err := NewTag(k, v)
	if err != nil {
		panic(err)
	}
	return t
}

// NewTag returns a tag that uses the provided key and value. The returned tag is normalized to conform with the DataDog
// tag specification. The key and value must be non-empty and the key must begin with a letter. The string form of the
// returned tag is "normalized(k):normalized(v)".
func NewTag(k, v string) (Tag, error) {
	if k == "" {
		return Tag{}, errors.New("key cannot be empty")
	}
	if v == "" {
		return Tag{}, errors.New("value cannot be empty")
	}

	k = strings.ToLower(k)
	if !(k[0] >= 'a' && k[0] <= 'z') {
		return Tag{}, errors.New("tag must start with a letter")
	}

	for _, r := range k {
		if r == ':' {
			return Tag{}, errors.New("tag key cannot contain ':'")
		}
	}

	// full tag, which is "key:value", must be <= 200 characters
	if tagLen := len(k) + 1 + len(v); tagLen > 200 {
		return Tag{}, errors.New(`full tag ("key:value") must be <= 200 characters`)
	}

	return Tag{
		key:   normalizeTag(k),
		value: normalizeTag(v),
	}, nil
}

// MustNewTags returns the result of calling NewTags, but panics if NewTags returns an error. Should only be used in
// instances where the inputs are statically defined and known to be valid.
func MustNewTags(t map[string]string) Tags {
	tags, err := NewTags(t)
	if err != nil {
		panic(err)
	}
	return tags
}

// NewTags returns a slice of tags that use the provided key:value mapping.
func NewTags(t map[string]string) (Tags, error) {
	var tags Tags
	for k, v := range t {
		tag, err := NewTag(k, v)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

var validChars = make(map[rune]struct{})

func init() {
	for ch := 'a'; ch <= 'z'; ch++ {
		validChars[ch] = struct{}{}
	}
	for ch := '0'; ch <= '9'; ch++ {
		validChars[ch] = struct{}{}
	}
	validChars['_'] = struct{}{}
	validChars['-'] = struct{}{}
	validChars[':'] = struct{}{}
	validChars['.'] = struct{}{}
	validChars['/'] = struct{}{}
}

// normalizeTag takes the given input string and normalizes it using the same rules as DataDog (https://help.datadoghq.com/hc/en-us/articles/204312749-Getting-started-with-tags):
// "Tags must start with a letter, and after that may contain alphanumerics, underscores, minuses, colons, periods and
// slashes. Other characters will get converted to underscores. Tags can be up to 200 characters long and support
// unicode. Tags will be converted to lowercase."
//
// Note that this function does not impose the length restriction described above.
func normalizeTag(in string) string {
	in = strings.ToLower(in)
	var out string
	for _, r := range in {
		if _, ok := validChars[r]; !ok {
			r = '_'
		}
		out += string(r)
	}
	return out
}