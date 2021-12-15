package meta

type TrackInfo struct {
	Title  string   `json:"title"`
	Artist string   `json:"artist"`
	Type   string   `json:"type"`
	Tags   []string `json:"tags"`
}

type TrackInfoWithAlbum struct {
	Title   string   `json:"title"`
	Artist  string   `json:"artist"`
	Type    string   `json:"type"`
	Tags    []string `json:"tags"`
	TrackID int      `json:"track_id"`
	DiscID  int      `json:"disc_id"`
	AlbumID string   `json:"album_id"`
}

type DiscInfo struct {
	Title   string       `json:"title"`
	Artist  string       `json:"artist"`
	Catalog string       `json:"catalog"`
	Tags    []string     `json:"tags"`
	Type    string       `json:"type"`
	Tracks  []*TrackInfo `json:"tracks"`
}

type AlbumInfo struct {
	AlbumID string      `json:"album_id"`
	Title   string      `json:"title"`
	Edition string      `json:"edition"`
	Catalog string      `json:"catalog"`
	Artist  string      `json:"artist"`
	Date    interface{} `json:"date"`
	Tags    []string    `json:"tags"`
	Type    string      `json:"type"`
	Discs   []*DiscInfo `json:"discs"`
}

type AlbumMeta struct {
	AlbumID string      `json:"album_id" toml:"album_id"`
	Title   string      `json:"title"`
	Edition string      `json:"edition"`
	Catalog string      `json:"catalog"`
	Artist  string      `json:"artist"`
	Date    interface{} `json:"date"`
	Tags    []string    `json:"tags"`
	Type    string      `json:"type"`
}

type Tag struct {
	Name       string   `json:"name"`
	Includes   []string `json:"includes"`
	IncludedBy []string `json:"included-by" toml:"included-by"`
}

type Record struct {
	Album AlbumMeta   `json:"album"`
	Discs []*DiscInfo `json:"discs"`
}
