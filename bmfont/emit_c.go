package bmfont

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

// Emits C code from parsed bmfont description.
// Ignores write errors
func (f *Font) EmitCFont(b io.StringWriter) {
	b.WriteString(fmt.Sprintf(header, f.Name))
	b.WriteString(types)

	// Emit variable declarations
	b.WriteString(fmt.Sprintf("const BMFontBucket BMFONT_%s_BUCKETS[];\n", f.Name))
	b.WriteString(fmt.Sprintf("const BMFontPage BMFONT_%s_PAGES[];\n", f.Name))
	for i := range f.Buckets {
		b.WriteString(fmt.Sprintf("const BMFontChar BMFONT_%s_BUCKET_%d[];\n", f.Name, i))
	}
	for i := range f.Pages {
		b.WriteString(fmt.Sprintf("unsigned char BMFONT_%s_PAGE_%d[];\n", f.Name, i))
	}

	// Emit font struct
	b.WriteString(fmt.Sprintf(`
const struct BMFont BMFONT_%s = {
    .size = %d,
    .lineHeight = %d,
	.base = %d,
	.scaleW = %d,
	.scaleH = %d,
	.isPacked = %d,
	.aChannelType = %s,
	.rChannelType = %s,
	.gChannelType = %s,
	.bChannelType = %s,

	.bucketCount = %d,
	.buckets = %s,
	.pageCount = %d,
	.pages = %s,
};
`,
		f.Name, f.Size, f.LineHeight, f.Base, f.ScaleW, f.ScaleH, f.IsPacked,
		f.AType, f.RType, f.GType, f.BType,
		len(f.Buckets), fmt.Sprintf("BMFONT_%s_BUCKETS", f.Name),
		len(f.Pages), fmt.Sprintf("BMFONT_%s_PAGES", f.Name),
	))

	// Emit buckets
	b.WriteString(fmt.Sprintf("\nconst BMFontBucket BMFONT_%s_BUCKETS[] = {\n", f.Name))
	for i, bucket := range f.Buckets {
		b.WriteString(fmt.Sprintf("    {%d, %d, %s},\n",
			bucket.StartChar, bucket.EndChar,
			fmt.Sprintf("BMFONT_%s_BUCKET_%d", f.Name, i),
		))
	}
	b.WriteString("};\n")

	// Emit pages
	b.WriteString(fmt.Sprintf("const BMFontPage BMFONT_%s_PAGES[] = {\n", f.Name))
	for i, page := range f.Pages {
		b.WriteString(fmt.Sprintf("    {%d, %s}, \n", page.PageSize,
			fmt.Sprintf("BMFONT_%s_PAGE_%d", f.Name, i),
		))
	}
	b.WriteString("};\n")

	// Emit kerning pairs
	for _, bucket := range f.Buckets {
		for _, char := range bucket.Chars {
			if len(char.Kernings) > 0 {
				b.WriteString(fmt.Sprintf("const BMFontKerning BMFONT_%s_KERNINGS_CHAR_%d[] = { // '%[2]c' \n", f.Name, char.ID))
				for secondChar, amount := range char.Kernings {
					b.WriteString(fmt.Sprintf("    {%d, %d}, // '%[1]c' \n", secondChar, amount))
				}
				b.WriteString("};\n\n")
			}
		}
	}

	// Emit chars
	kernings := []Char{} // chars with kerning pairs
	for i, bucket := range f.Buckets {
		b.WriteString(fmt.Sprintf("const BMFontChar BMFONT_%s_BUCKET_%d[] = {\n", f.Name, i))
		for _, char := range bucket.Chars {
			kerningPage := "NULL"
			if len(char.Kernings) > 0 {
				kerningPage = fmt.Sprintf("BMFONT_%s_KERNINGS_CHAR_%d", f.Name, char.ID)
			}
			b.WriteString(fmt.Sprintf("    {%d, %d, %d, %d, %d, %d, %d, %d, %d, %d, %s}, // %d ('%[12]c')\n",
				char.X, char.Y, char.Width, char.Height, char.XOffset, char.YOffset,
				char.XAdvance, char.Page, len(char.Kernings), char.Channels,
				kerningPage,
				char.ID,
			))
			if len(char.Kernings) > 0 {
				kernings = append(kernings, char)
			}
		}
		b.WriteString("};\n\n")
	}

	// Emit page data
	for i, page := range f.Pages {
		b.WriteString(fmt.Sprintf("unsigned char BMFONT_%s_PAGE_%d[] __attribute__((aligned(16))) = {\n    ", f.Name, i))

		scanner := bufio.NewScanner(page.PageData)
		scanner.Split(bufio.ScanBytes)
		bytesWritten := 0
		for scanner.Scan() {
			b.WriteString(fmt.Sprintf("0x%x, ", scanner.Bytes()))
			bytesWritten++
			if bytesWritten == 16 {
				b.WriteString("\n    ")
				bytesWritten = 0
			}
		}
		page.PageData.Close()
		if err := scanner.Err(); err != nil {
			log.Fatalf("failed to read file: %s", err)
		}
		b.WriteString("\n};\n\n")
	}

	b.WriteString(footer)
}

const header = `#ifndef _%s_H_
#define _%[1]s_H_
`

const types = `

#ifndef _BMFONT_TYPES_
#define _BMFONT_TYPES_

#include <stdint.h>
#include <stddef.h>

typedef struct BMFontKerning {
  uint32_t secondChar;
  int16_t amount;
} BMFontKerning;

typedef struct BMFontChar {
  uint16_t x;
  uint16_t y;
  uint16_t width;
  uint16_t height;
  int16_t xoffset;
  int16_t yoffset;
  int16_t xadvance;
  uint8_t page;
  uint8_t kerningsCount;
  uint8_t channels; // 1 — blue, 2 — green, 4 — red, 15 — all
  const BMFontKerning *kernings;
} BMFontChar;

typedef struct BMFontBucket {
  uint32_t startChar; // char with number ch can be found using idx+startChar
  uint32_t endChar;
  const BMFontChar *chars;
} BMFontBucket;

typedef struct BMFontPage {
  unsigned int size;
  unsigned char *data; // PNG
} BMFontPage;

// Enum for supported modes
typedef enum {
  CHANNEL_GLYPH,
  CHANNEL_OUTLINE,
  CHANNEL_GLYPH_OUTLINE,
  CHANNEL_ZERO,
  CHANNEL_ONE,
} ChannelType;

typedef struct BMFont {
  uint16_t scaleW;
  uint16_t scaleH;
  uint16_t size;
  uint8_t lineHeight;
  uint8_t base;
  uint8_t isPacked;
  ChannelType aChannelType;
  ChannelType rChannelType;
  ChannelType gChannelType;
  ChannelType bChannelType;

  uint16_t bucketCount;
  const BMFontBucket *buckets;
  uint16_t pageCount;
  const BMFontPage *pages;
} BMFont;

#endif

`
const footer = `#endif`
