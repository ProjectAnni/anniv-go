package meta

type TrackInfo struct {
	Title   string   `json:"title" toml:"title"`
	Artist  string   `json:"artist" toml:"artist"`
	Artists *Artists `json:"artists" toml:"artists"`
	Type    string   `json:"type" toml:"type"`
	Tags    []string `json:"tags" toml:"tags"`
}

type TrackInfoWithAlbum struct {
	Title      string   `json:"title" toml:"title" mapstructure:"title"`
	Artist     string   `json:"artist" toml:"artist" mapstructure:"artist"`
	Type       string   `json:"type" toml:"type" mapstructure:"type"`
	Tags       []string `json:"tags" toml:"tags" mapstructure:"tags"`
	TrackID    int      `json:"track_id" toml:"track_id" mapstructure:"track_id"`
	DiscID     int      `json:"disc_id" toml:"disc_id" mapstructure:"disc_id"`
	AlbumID    string   `json:"album_id" toml:"album_id" mapstructure:"album_id"`
	AlbumTitle string   `json:"album_title" toml:"album_title" mapstructure:"album_title"`
}

type DiscInfo struct {
	Title   string       `json:"title" toml:"title"`
	Artist  string       `json:"artist" toml:"artist"`
	Artists *Artists     `json:"artists" toml:"artists"`
	Catalog string       `json:"catalog" toml:"catalog"`
	Tags    []string     `json:"tags" toml:"tags"`
	Type    string       `json:"type" toml:"type"`
	Tracks  []*TrackInfo `json:"tracks" toml:"tracks"`
}

type AlbumInfo struct {
	AlbumID string      `json:"album_id" toml:"album_id"`
	Title   string      `json:"title" toml:"title"`
	Edition *string     `json:"edition" toml:"edition"`
	Catalog string      `json:"catalog" toml:"catalog"`
	Artist  string      `json:"artist" toml:"artist"`
	Artists *Artists    `json:"artists" toml:"artists"`
	Date    interface{} `json:"date" toml:"date"`
	Tags    []string    `json:"tags" toml:"tags"`
	Type    string      `json:"type" toml:"type"`
	Discs   []*DiscInfo `json:"discs" toml:"discs"`
}

type AlbumMeta struct {
	AlbumID string      `json:"album_id" toml:"album_id"`
	Title   string      `json:"title" toml:"title"`
	Edition *string     `json:"edition" toml:"edition"`
	Catalog string      `json:"catalog" toml:"catalog"`
	Artist  string      `json:"artist" toml:"artist"`
	Artists *Artists    `json:"artists" toml:"artists"`
	Date    interface{} `json:"date" toml:"date"`
	Tags    []string    `json:"tags" toml:"tags"`
	Type    string      `json:"type" toml:"type"`
}

type Tag struct {
	Name  string   `json:"name" toml:"name"`
	Type  string   `json:"type" toml:"type"`
	Alias []string `json:"alias" toml:"alias"`
}

type TagDef struct {
	Tag
	Includes   map[string][]string `json:"includes" toml:"includes"`
	IncludedBy []string            `json:"included-by" toml:"included-by"`
}

type Record struct {
	Album AlbumMeta   `json:"album" toml:"album"`
	Discs []*DiscInfo `json:"discs" toml:"discs"`
}

type Artists map[string]string
