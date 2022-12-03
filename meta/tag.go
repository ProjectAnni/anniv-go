package meta

import (
	"errors"
	"github.com/pelletier/go-toml/v2"
	"os"
	"path"
)

type TagRef interface{}

var tagIdx map[string][]AlbumDetails
var tagIdxNonRecursive map[string][]AlbumDetails
var tags []Tag
var tagNameIdx map[string][]Tag
var tagGraph map[string][]string

func checkTags(V []Tag, E map[string][]string) error {
	VSet := make(map[string]int)
	for _, v := range V {
		VSet[v.Name]++
	}
	// Node constraints check
	if len(VSet) != len(V) {
		msg := "duplicated tags detected: "
		first := true
		for k, v := range VSet {
			if v != 1 {
				if !first {
					msg += ","
				} else {
					first = false
				}
				msg += k
			}
		}
		return errors.New(msg)
	}
	for k, v := range E {
		_, e := VSet[k]
		if !e {
			return errors.New("unknown tag: " + k)
		}
		for _, t := range v {
			_, e := VSet[t]
			if !e {
				return errors.New("unknown tag: " + t)
			}
		}
	}
	acc := make(map[string]bool)
	vis := make(map[string]bool)
	// Loop detect
	var dfs func(string) error
	dfs = func(x string) error {
		acc[x] = true
		vis[x] = true
		for _, nx := range E[x] {
			if acc[nx] {
				return errors.New("loop detected")
			}
			if !vis[nx] {
				err := dfs(nx)
				if err != nil {
					return err
				}
			}
		}
		acc[x] = false
		return nil
	}

	for _, v := range V {
		if !vis[v.Name] {
			err := dfs(v.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func readTags(p string) ([]Tag, map[string][]string, error) {
	E := make(map[string][]string)
	V := make([]Tag, 0)

	f, err := os.ReadDir(p)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range f {
		f, err := os.Open(path.Join(p, v.Name()))
		if err != nil {
			return nil, nil, errors.New(v.Name() + ": " + err.Error())
		}
		tags := make(map[string][]tagDef, 0)
		err = toml.NewDecoder(f).Decode(&tags)
		if err != nil {
			return nil, nil, errors.New(v.Name() + ": " + err.Error())
		}
		f.Close()
		for _, tag := range tags["tag"] {
			if !isValidTagType(tag.Type) {
				return nil, nil, errors.New(v.Name() + ": " + "invalid tag type: " + tag.Type)
			}
			V = append(V, Tag{
				Name:  tag.Name,
				Type:  tag.Type,
				Names: tag.Names,
			})
			for typ, child := range tag.Includes {
				if !isValidTagType(typ) {
					return nil, nil, errors.New(v.Name() + ": " + "invalid tag type: " + typ)
				}
				for _, v := range child {
					V = append(V, Tag{
						Name: v,
						Type: typ,
					})
					E[tag.Name] = append(E[tag.Name], v)
				}
			}
			for _, parent := range tag.IncludedBy {
				E[parent] = append(E[parent], tag.Name)
			}
		}
	}
	return V, E, nil
}

var validTypes = []string{"artist", "group", "animation", "series", "project", "game", "organization", "default", "category"}

func isValidTagType(p string) bool {
	for _, v := range validTypes {
		if p == v {
			return true
		}
	}
	return false
}

var ErrTagNotFound = errors.New("tag not found")
var ErrTagAmbiguous = errors.New("ambiguous tag")
var ErrInvalidTagRef = errors.New("invalid tag ref")

func QueryTag(ref TagRef) (*Tag, error) {
	var err error
	if tagName, ok := ref.(string); ok {
		arr := tagNameIdx[tagName]
		if len(arr) == 0 {
			return nil, ErrTagNotFound
		}
		if len(arr) > 1 {
			return nil, ErrTagAmbiguous
		}
		return &arr[0], nil
	}

	if tagArray, ok := ref.([]string); ok {
		if len(tagArray) == 0 {
			return nil, ErrInvalidTagRef
		}
		if len(tagArray) == 1 {
			return QueryTag(tagArray[0])
		}
		var maybeTag []*Tag
		for _, v := range tagNameIdx[tagArray[0]] {
			maybeTag = append(maybeTag, &v)
		}

		var parentTag *Tag

		if isValidTagType(tagArray[1]) {
			var tmpMaybeTag []*Tag
			for _, v := range maybeTag {
				if v.Type == tagArray[1] {
					tmpMaybeTag = append(tmpMaybeTag, v)
				}
			}
			maybeTag = tmpMaybeTag
			parentTag, err = QueryTag(tagArray[2:])
			if err != nil {
				return nil, err
			}
		} else {
			parentTag, err = QueryTag(tagArray[1:])
			if err != nil {
				return nil, err
			}
		}

		var res []*Tag
		for _, v := range maybeTag {
			if v.parent == parentTag {
				res = append(res, v)
			}
		}

		if len(res) == 0 {
			return nil, ErrTagNotFound
		}
		if len(res) > 1 {
			return nil, ErrTagAmbiguous
		}
		return res[0], nil
	}

	return nil, ErrInvalidTagRef
}
