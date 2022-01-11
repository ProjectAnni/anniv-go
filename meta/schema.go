package meta

type TrackInfo struct {
	Title   string   `json:"title" toml:"title"`
	Artist  string   `json:"artist" toml:"artist"`
	Artists *Artists `json:"artists" toml:"artists"`
	Type    string   `json:"type" toml:"type"`
	Tags    []string `json:"tags" toml:"tags"`
}

type TrackInfoWithAlbum struct {
	Title      string   `json:"title" toml:"title"`
	Artist     string   `json:"artist" toml:"artist"`
	Type       string   `json:"type" toml:"type"`
	Tags       []string `json:"tags" toml:"tags"`
	TrackID    int      `json:"track_id" toml:"track_id"`
	DiscID     int      `json:"disc_id" toml:"disc_id"`
	AlbumID    string   `json:"album_id" toml:"album_id"`
	AlbumTitle string   `json:"album_title" toml:"album_title"`
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
	Name       string   `json:"name" toml:"name"`
	Includes   []string `json:"includes" toml:"includes"`
	IncludedBy []string `json:"included-by" toml:"included-by"`
}

type Record struct {
	Album AlbumMeta   `json:"album" toml:"album"`
	Discs []*DiscInfo `json:"discs" toml:"discs"`
}

type Artists struct {
	Vocal     *string `json:"vocal" toml:"vocal"`
	Composer  *string `json:"composer" toml:"composer"`
	Lyricist  *string `json:"lyricist" toml:"lyricist"`
	Arranger  *string `json:"arranger" toml:"arranger"`
	Piano     *string `json:"piano" toml:"piano"`
	Violin    *string `json:"violin" toml:"violin"`
	Viola     *string `json:"viola" toml:"viola"`
	Cello     *string `json:"cello" toml:"cello"`
	IrishHarp *string `json:"irish_harp" toml:"irish-harp"`
}
