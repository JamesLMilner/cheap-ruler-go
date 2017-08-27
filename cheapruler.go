package cheapruler

import (
	"errors"
	"math"
)

// A collection of very fast approximations to common geodesic measurements.
// Useful for performance-sensitive code that measures things on a city scale.
type CheapRuler struct {
	Kx      float64
	Ky      float64
	Factors map[string]float64
}

// The closest point on the line from the given point and
// index is the start index of the segment with the closest point.
type PointOnLine struct {
	Point []float64
	Index float64
	T     float64
}

// Create a new cheap ruler instance
func NewCheapruler(lat float64, units string) (CheapRuler, error) {

	cr := CheapRuler{}

	// Multipliers for converting between units.
	factors := map[string]float64{
		"kilometers":    1,
		"miles":         1000 / 1609.344,
		"nauticalmiles": 1000 / 1852,
		"meters":        1000,
		"metres":        1000,
		"yards":         1000 / 0.9144,
		"feet":          1000 / 0.3048,
		"inches":        1000 / 0.0254,
	}

	if m, ok := factors[units]; ok {

		cos := math.Cos(lat * math.Pi / 180)
		cos2 := 2*cos*cos - 1
		cos3 := 2*cos*cos2 - cos
		cos4 := 2*cos*cos3 - cos2
		cos5 := 2*cos*cos4 - cos3

		// multipliers for converting longitude and latitude degrees into distance
		// (http://1.usa.gov/1Wb1bv7)
		cr.Kx = m * (111.41513*cos - 0.09455*cos3 + 0.00012*cos5)
		cr.Ky = m * (111.13209 - 0.56605*cos2 + 0.0012*cos4)
		cr.Factors = factors

		return cr, nil

	} else {

		err := errors.New(units + "is not a valid unit")
		return cr, err

	}

}

// Creates a CheapRuler struct from tile coordinates (y and z). Convenient in tile-reduce scripts.
func NewCheaprulerFromTile(y float64, z float64, units string) (CheapRuler, error) {
	n := math.Pi * (1 - 2*(y+0.5)/math.Pow(2, z))
	lat := math.Atan(0.5*(math.Exp(n)-math.Exp(-n))) * 180 / math.Pi
	return NewCheapruler(lat, units)
}

// Given two points returns the distance in the units of the ruler
func (cr CheapRuler) Distance(a []float64, b []float64) float64 {
	dx := (a[0] - b[0]) * cr.Kx
	dy := (a[1] - b[1]) * cr.Ky
	return math.Sqrt(dx*dx + dy*dy)
}

// Returns the bearing between two points in angles.
func (cr CheapRuler) Bearing(a []float64, b []float64) float64 {
	dx := (b[0] - a[0]) * cr.Kx
	dy := (b[1] - a[1]) * cr.Ky
	if dx == 0.0 && dy == 0.0 {
		return 0.0
	}
	bearing := math.Atan2(dx, dy) * 180 / math.Pi
	if bearing > 180 {
		bearing -= 360
	}
	return bearing
}

// Returns a new point given distance and bearing from the starting point.
func (cr CheapRuler) Destination(p []float64, dist float64, bearing float64) []float64 {
	a := (90.0 - bearing) * math.Pi / 180.0
	return cr.Offset(p, math.Cos(a)*dist, math.Sin(a)*dist)
}

// Returns a new point given easting and northing offsets (in ruler units) from the starting point.
func (cr CheapRuler) Offset(p []float64, dx float64, dy float64) []float64 {
	xo := p[0] + dx/cr.Kx
	yo := p[1] + dy/cr.Ky
	return []float64{xo, yo}
}

// Given a line (an slice of points), returns the total line distance.
func (cr CheapRuler) LineDistance(points [][]float64) float64 {
	total := 0.0
	for i := 0; i < len(points)-1; i++ {
		total += cr.Distance(points[i], points[i+1])
	}
	return total
}

// Given a polygon (a slice of rings, where each ring is a slice of points), returns the area.
func (cr CheapRuler) Area(polygon [][][]float64) float64 {
	sum := 0.0

	for i := 0; i < len(polygon); i++ {
		ring := polygon[i]
		ringlen := len(ring)
		k := ringlen - 1.0

		for j := 0; j < ringlen; {
			posneg := 1.0
			if i != 0 {
				posneg = -1.0
			}
			sum += (ring[j][0] - ring[k][0]) * (ring[j][1] + ring[k][1]) * posneg

			j++
			k = j
		}
	}

	return (math.Abs(sum) / 2) * cr.Kx * cr.Ky
}

// Returns the point at a specified distance along the line.
func (cr CheapRuler) Along(line [][]float64, dist float64) []float64 {
	sum := 0.0

	if dist <= 0 {
		return line[0]
	}

	for i := 0; i < len(line)-1; i++ {
		p0 := line[i]
		p1 := line[i+1]
		d := cr.Distance(p0, p1)
		sum += d
		if sum > dist {
			return interpolate(p0, p1, (dist-(sum-d))/d)
		}
	}

	return line[len(line)-1]
}

// Returns an struct where point is closest point on the line from the given point,
// and index is the start index of the segment with the closest point.
func (cr CheapRuler) PointOnLine(line [][]float64, p []float64) PointOnLine {
	minDist := math.Inf(1)
	var minX float64
	var minY float64
	var minI float64
	var minT float64
	var t float64

	for i := 0; i < len(line)-1; i++ {

		x := line[i][0]
		y := line[i][1]
		dx := (line[i+1][0] - x) * cr.Kx
		dy := (line[i+1][1] - y) * cr.Ky

		if dx != 0 || dy != 0 {

			t = ((p[0]-x)*cr.Kx*dx + (p[1]-y)*cr.Ky*dy) / (dx*dx + dy*dy)

			if t > 1 {
				x = line[i+1][0]
				y = line[i+1][1]

			} else if t > 0 {
				x += (dx / cr.Kx) * t
				y += (dy / cr.Ky) * t
			}
		}

		dx = (p[0] - x) * cr.Kx
		dy = (p[1] - y) * cr.Ky

		sqDist := dx*dx + dy*dy
		if sqDist < minDist {
			minDist = sqDist
			minX = x
			minY = y
			minI = float64(i)
			minT = t
		}
	}

	return PointOnLine{
		[]float64{minX, minY},
		minI,
		minT,
	}
}

// Returns a part of the given line between the start and the stop points (or their closest points on the line).
func (cr CheapRuler) LineSlice(start []float64, stop []float64, line [][]float64) [][]float64 {
	p1 := cr.PointOnLine(line, start)
	p2 := cr.PointOnLine(line, stop)

	if p1.Index > p2.Index || (p1.Index == p2.Index && p1.T > p2.T) {
		tmp := p1
		p1 = p2
		p2 = tmp
	}

	sl := [][]float64{p1.Point}

	l := p1.Index + 1
	r := p2.Index

	if !equals(line[int(l)], sl[0]) && l <= r {
		sl = append(sl, line[int(l)])
	}

	for i := l + 1; i <= r; i++ {
		sl = append(sl, line[int(i)])
	}

	if !equals(line[int(r)], p2.Point) {
		sl = append(sl, p2.Point)
	}

	return sl
}

// Returns a part of the given line between the start and the stop points indicated by distance along the line.
func (cr CheapRuler) LineSliceAlong(start float64, stop float64, line [][]float64) [][]float64 {
	sum := 0.0
	var sl [][]float64

	for i := 0; i < len(line)-1; i++ {
		p0 := line[i]
		p1 := line[i+1]
		d := cr.Distance(p0, p1)

		sum += d

		if sum > start && len(sl) == 0.0 {
			sl = append(sl, interpolate(p0, p1, (start-(sum-d))/d))
		}

		if sum >= stop {
			sl = append(sl, interpolate(p0, p1, (stop-(sum-d))/d))
			return sl
		}

		if sum > start {
			sl = append(sl, p1)
		}
	}

	return sl
}

// Given a point, returns a bounding box slice ([]float64{w, s, e, n})
// created from the given point buffered by a given distance.
func (cr CheapRuler) BufferPoint(p []float64, buffer float64) []float64 {
	v := buffer / cr.Ky
	h := buffer / cr.Kx
	return []float64{
		p[0] - h,
		p[1] - v,
		p[0] + h,
		p[1] + v,
	}
}

// Given a bounding box, returns the box buffered by a given distance.
func (cr CheapRuler) BufferBBox(bbox []float64, buffer float64) []float64 {
	v := buffer / cr.Ky
	h := buffer / cr.Kx
	return []float64{
		bbox[0] - h,
		bbox[1] - v,
		bbox[2] + h,
		bbox[3] + v,
	}
}

// Returns true if the given point is inside in the given bounding box, otherwise false.
func (cr CheapRuler) InsideBBox(p []float64, bbox []float64) bool {
	return p[0] >= bbox[0] &&
		p[0] <= bbox[2] &&
		p[1] >= bbox[1] &&
		p[1] <= bbox[3]
}

func equals(a []float64, b []float64) bool {
	return a[0] == b[0] && a[1] == b[1]
}

func interpolate(a []float64, b []float64, t float64) []float64 {
	dx := b[0] - a[0]
	dy := b[1] - a[1]
	return []float64{
		a[0] + dx*t,
		a[1] + dy*t,
	}
}
