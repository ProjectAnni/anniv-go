package meta

type Artists map[string]string

type AlbumIdentifier string

type DiscIdentifier struct {
	AlbumID AlbumIdentifier `json:"album_id"  mapstructure:"albumID"`
	DiscID  uint            `json:"disc_id" mapstructure:"discID"`
}

type TrackIdentifier struct {
	DiscIdentifier
	TrackID uint `json:"track_id" mapstructure:"track_id"`
}

type AlbumInfo struct {
	AlbumID AlbumIdentifier `json:"album_id" toml:"album_id"`
	Title   string          `json:"title" toml:"title"`
	Edition *string         `json:"edition,omitempty" toml:"edition"`
	Catalog string          `json:"catalog" toml:"catalog"`
	Artist  string          `json:"artist" toml:"artist"`
	Date    string          `json:"date" toml:"date"`
	Type    string          `json:"type" toml:"type"`
}

type AlbumDetails struct {
	AlbumInfo
	Artists *Artists       `json:"artists,omitempty" toml:"artists"`
	Tags    []string       `json:"tags" toml:"tags"`
	Discs   []*DiscDetails `json:"discs" toml:"discs"`
}

type albumInfoDef struct {
	AlbumDetails
	Date interface{} `json:"date" toml:"date"`
}

type DiscInfo struct {
	Title   *string `json:"title,omitempty" toml:"title"`
	Artist  *string `json:"artist,omitempty" toml:"artist"`
	Catalog string  `json:"catalog" toml:"catalog"`
	Type    *string `json:"type,omitempty" toml:"type"`
}

type DiscDetails struct {
	DiscInfo
	Artists *Artists       `json:"artists,omitempty" toml:"artists"`
	Tags    []string       `json:"tags,omitempty" toml:"tags"`
	Tracks  []*TrackDetail `json:"tracks" toml:"tracks"`
}

type TrackInfo struct {
	Title  string  `json:"title" toml:"title"`
	Artist *string `json:"artist,omitempty" toml:"artist"`
	Type   *string `json:"type,omitempty" toml:"type"`
}

type TrackDetail struct {
	TrackInfo
	Artists *Artists `json:"artists,omitempty" toml:"artists"`
	Tags    []string `json:"tags,omitempty" toml:"tags"`
}

type Tag struct {
	Name  string   `json:"name" toml:"name"`
	Type  string   `json:"type" toml:"type"`
	Alias []string `json:"alias" toml:"alias"`
}

type TrackInfoWithAlbum struct {
	TrackIdentifier
	TrackInfo
	AlbumTitle string `json:"album_title"`
}

type tagDef struct {
	Tag
	Includes   map[string][]string `json:"includes" toml:"includes"`
	IncludedBy []string            `json:"included-by" toml:"included-by"`
}

type record struct {
	Album albumInfoDef   `json:"album" toml:"album"`
	Discs []*DiscDetails `json:"discs" toml:"discs"`
}
