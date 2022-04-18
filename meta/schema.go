package meta

type TrackInfo struct {
	Title  string `json:"title" toml:"title"`
	Artist string `json:"artist" toml:"artist"`
	Type   string `json:"type" toml:"type"`
}

type TrackDetail struct {
	TrackInfo
	Artists *Artists `json:"artists" toml:"artists"`
	Tags    []string `json:"tags" toml:"tags"`
}

type DiscInfo struct {
	Title   string `json:"title" toml:"title"`
	Artist  string `json:"artist" toml:"artist"`
	Catalog string `json:"catalog" toml:"catalog"`
	Type    string `json:"type" toml:"type"`
}

type DiscDetails struct {
	DiscInfo
	Artists *Artists       `json:"artists" toml:"artists"`
	Tags    []string       `json:"tags" toml:"tags"`
	Tracks  []*TrackDetail `json:"tracks" toml:"tracks"`
}

type AlbumInfo struct {
	AlbumID string  `json:"album_id" toml:"album_id"`
	Title   string  `json:"title" toml:"title"`
	Edition *string `json:"edition" toml:"edition"`
	Catalog string  `json:"catalog" toml:"catalog"`
	Artist  string  `json:"artist" toml:"artist"`
	Date    string  `json:"date" toml:"date"`
	Type    string  `json:"type" toml:"type"`
}

type AlbumDetails struct {
	AlbumInfo
	Artists *Artists       `json:"artists" toml:"artists"`
	Tags    []string       `json:"tags" toml:"tags"`
	Discs   []*DiscDetails `json:"discs" toml:"discs"`
}

type Tag struct {
	Name  string   `json:"name" toml:"name"`
	Type  string   `json:"type" toml:"type"`
	Alias []string `json:"alias" toml:"alias"`
}

type tagDef struct {
	Tag
	Includes   map[string][]string `json:"includes" toml:"includes"`
	IncludedBy []string            `json:"included-by" toml:"included-by"`
}

type record struct {
	Album AlbumDetails   `json:"album" toml:"album"`
	Discs []*DiscDetails `json:"discs" toml:"discs"`
}

type Artists map[string]string
