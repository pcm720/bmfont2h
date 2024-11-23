package bmfont

import (
	"os"
)

// Defines basic font metadata
type Font struct {
	Name           string
	Size           uint32
	LineHeight     uint32
	AType          ChannelType
	RType          ChannelType
	GType          ChannelType
	BType          ChannelType
	ScaleH, ScaleW uint16
	Base           uint8
	IsPacked       uint8

	Buckets []Bucket
	Pages   []Page // Contains opened os.File references
}

// Defines char metadata
type Char struct {
	ID                         uint32
	X, Y, Width, Height        uint16
	XOffset, YOffset, XAdvance int16
	Page, Channels             uint8
	Kernings                   map[uint32]int16 // second char - amount
}

// Defines a single bucket that contains consecutive chars
type Bucket struct {
	StartChar, EndChar uint32
	Chars              []Char
}

// Defines a single page
type Page struct {
	PageSize uint32
	PageData *os.File // Must be closed after use
}

// Defines channel types
type ChannelType string

const (
	ChannelType_Glyph        = "CHANNEL_GLYPH"
	ChannelType_Outline      = "CHANNEL_OUTLINE"
	ChannelType_GlyphOutline = "CHANNEL_GLYPH_OUTLINE"
	ChannelType_Zero         = "CHANNEL_ZERO"
	ChannelType_One          = "CHANNEL_ONE"
)

// Maps bmfont channel index to ChannelType
var ChannelTypeMapping = []ChannelType{
	ChannelType_Glyph,
	ChannelType_Outline,
	ChannelType_GlyphOutline,
	ChannelType_Zero,
	ChannelType_One,
}
