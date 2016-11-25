package pixel

import "github.com/faiface/pixel/pixelgl"

// Warning: technical stuff below.

// VertexFormat is an internal format of the OpenGL vertex data.
//
// You can actually change this and all of the Pixel's functions will use the new format.
// Only change when you're implementing an OpenGL effect or something similar.
var VertexFormat = DefaultVertexFormat()

// DefaultVertexFormat returns the default vertex format used by Pixel.
func DefaultVertexFormat() pixelgl.VertexFormat {
	return pixelgl.VertexFormat{
		{Purpose: pixelgl.Position, Size: 2},
		{Purpose: pixelgl.Color, Size: 4},
		{Purpose: pixelgl.TexCoord, Size: 2},
	}
}

// ConvertVertexData converts data in the oldFormat to the newFormat. Vertex attributes in the new format
// will be copied from the corresponding vertex attributes in the old format. If a vertex attribute in the new format
// has no corresponding attribute in the old format, it will be filled with zeros.
func ConvertVertexData(oldFormat, newFormat pixelgl.VertexFormat, data []float64) []float64 {
	// calculate the mapping between old and new format
	// if i is a start of a vertex attribute in the new format, then mapping[i] returns
	// the index where the same attribute starts in the old format
	mapping := make(map[int]int)
	i := 0
	for _, newAttr := range newFormat {
		j := 0
		for _, oldAttr := range oldFormat {
			if newAttr == oldAttr {
				mapping[i] = j
				break
			}
			j += oldAttr.Size
		}
		i += newAttr.Size
	}

	oldData, newData := data, []float64{}

	for i := 0; i < len(oldData); i += oldFormat.Size() {
		j := 0
		for _, attr := range newFormat {
			if oldIndex, ok := mapping[j]; ok { // the attribute was found in the old format
				newData = append(newData, oldData[i+oldIndex:i+oldIndex+attr.Size]...)
			} else { // the attribute wasn't found in the old format, so fill with zeros
				newData = append(newData, make([]float64, attr.Size)...)
			}
			j += attr.Size
		}
	}

	return newData
}
