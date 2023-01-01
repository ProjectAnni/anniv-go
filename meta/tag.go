package meta

import (
	"errors"
	"strings"
)

var (
	ErrTagDefAmbiguous = errors.New("ambiguous tag definition")
	ErrInvalidTagType  = errors.New("invalid tag type")
	ErrDuplicatedTags  = errors.New("duplicated tags")
	ErrUndefinedTag    = errors.New("undefined tag")
)

type Tag struct {
	Name                   string            `json:"name" toml:"name"`
	Type                   string            `json:"type" toml:"type"`
	Names                  map[string]string `json:"names" toml:"names"`
	parentTags             []string
	parentTagsRef          []*Tag
	childrenRef            []*Tag
	includeAlbums          map[*AlbumDetails]bool
	includeAlbumsRecursive map[*AlbumDetails]bool
}

func (tag *Tag) Str() string {
	return tag.Type + ":" + tag.Name
}

func (tag *Tag) addAlbum(album *AlbumDetails, direct bool) {
	if tag.includeAlbums == nil {
		tag.includeAlbums = map[*AlbumDetails]bool{}
	}
	if tag.includeAlbumsRecursive == nil {
		tag.includeAlbumsRecursive = map[*AlbumDetails]bool{}
	}

	if direct {
		tag.includeAlbums[album] = true
	}

	tag.includeAlbumsRecursive[album] = true

	for _, parent := range tag.parentTagsRef {
		parent.addAlbum(album, false)
	}
}

func (tag *Tag) AddAlbum(album *AlbumDetails) {
	tag.addAlbum(album, true)
}

func (tag *Tag) GetAlbums(recursive bool) []AlbumDetails {
	var res []AlbumDetails
	if recursive {
		for album := range tag.includeAlbumsRecursive {
			res = append(res, *album)
		}
	} else {
		for album := range tag.includeAlbums {
			res = append(res, *album)
		}
	}
	return res
}

type TagSet struct {
	tags       []Tag
	tagNameIdx map[string]int
	tagStrIdx  map[string]int
	tagGraph   map[string][]string
}

func (set *TagSet) FindTag(str string) (*Tag, error) {
	typ, name := ParseTagStr(str)
	if typ != nil {
		// find tag by both type and name
		formattedStr := *typ + ":" + name
		res, exist := set.tagStrIdx[formattedStr]
		if exist {
			return &set.tags[res], nil
		} else {
			return nil, ErrUndefinedTag
		}
	} else {
		// find tag only by name
		res, exist := set.tagNameIdx[name]
		if !exist {
			return nil, ErrUndefinedTag
		}
		if res == -1 {
			return nil, ErrTagDefAmbiguous
		}
		return &set.tags[res], nil
	}
}

func NewTagSet(tagsIn []Tag) (*TagSet, error) {
	set := TagSet{
		tags:       tagsIn,
		tagNameIdx: map[string]int{},
		tagStrIdx:  map[string]int{},
	}

	for _, tag := range tagsIn {
		if !validateTagType(tag.Type) {
			return nil, ErrInvalidTagType
		}
	}

	// build tag name idx
	for idx, tag := range tagsIn {
		_, exist := set.tagNameIdx[tag.Name]
		if exist {
			set.tagNameIdx[tag.Name] = -1
		} else {
			set.tagNameIdx[tag.Name] = idx
		}
	}

	// build tag str idx
	// at this step we ensured no
	// duplicated tags are present
	for idx, tag := range tagsIn {
		_, exist := set.tagStrIdx[tag.Str()]
		if exist {
			return nil, ErrDuplicatedTags
		} else {
			set.tagStrIdx[tag.Str()] = idx
		}
	}

	// expand tag refs
	for idx, tag := range set.tags {
		for _, parentStr := range tag.parentTags {
			parentTag, err := set.FindTag(parentStr)
			if err != nil {
				return nil, err
			}
			//tag.parentTagsRef = append(tag.parentTagsRef, parentTag)
			set.tags[idx].parentTagsRef = append(set.tags[idx].parentTagsRef, parentTag)
			parentTag.childrenRef = append(parentTag.childrenRef, &set.tags[idx])
		}
	}

	// TODO: loop check

	// build tag graph
	set.tagGraph = map[string][]string{}
	for _, tag := range tagsIn {
		tagStr := tag.Str()
		for _, nxt := range tag.childrenRef {
			set.tagGraph[tagStr] = append(set.tagGraph[tagStr], nxt.Str())
		}
	}

	return &set, nil
}

func ParseTagStr(str string) (*string, string) {
	idx := strings.Index(str, ":")
	if idx == -1 {
		return nil, strings.TrimSpace(str)
	}

	typ := str[:idx]
	name := strings.TrimSpace(str[idx+1:])

	if validateTagType(typ) {
		return &typ, name
	} else {
		return nil, strings.TrimSpace(str)
	}
}

func TagFromStr(str string) (*Tag, error) {
	typ, name := ParseTagStr(str)
	if typ == nil {
		return nil, ErrTagDefAmbiguous
	}
	if !validateTagType(*typ) {
		return nil, ErrInvalidTagType
	}
	tag := Tag{
		Name:  name,
		Type:  *typ,
		Names: nil,
	}
	return &tag, nil
}

var validTypes = []string{"artist", "group", "animation", "radio", "series", "project", "game", "organization", "unknown", "category"}

func validateTagType(p string) bool {
	for _, v := range validTypes {
		if p == v {
			return true
		}
	}
	return false
}
