package cheapruler

import "math"

type cheapruler struct {
	kx      float64
	ky      float64
	factors map[string]float64
}

type pol struct {
	point []float64
	index float64
	t     float64
}

func NewCheapruler(lat float64, units string) cheapruler {

	cr := cheapruler{}
	m := 1.0
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

	if units != "" {
		m = factors[units]
	}

	cos := math.Cos(lat * math.Pi / 180)
	cos2 := 2*cos*cos - 1
	cos3 := 2*cos*cos2 - cos
	cos4 := 2*cos*cos3 - cos2
	cos5 := 2*cos*cos4 - cos3

	// multipliers for converting longitude and latitude degrees into distance
	// (http://1.usa.gov/1Wb1bv7)
	cr.kx = m * (111.41513*cos - 0.09455*cos3 + 0.00012*cos5)
	cr.ky = m * (111.13209 - 0.56605*cos2 + 0.0012*cos4)
	cr.factors = factors
	return cr
}

func NewCheaprulerFromTile(y float64, z float64, units string) cheapruler {
	n := math.Pi * (1 - 2*(y+0.5)/math.Pow(2, z))
	lat := math.Atan(0.5*(math.Exp(n)-math.Exp(-n))) * 180 / math.Pi
	return NewCheapruler(lat, units)
}

func (cr cheapruler) distance(a []float64, b []float64) float64 {
	dx := (a[0] - b[0]) * cr.kx
	dy := (a[1] - b[1]) * cr.ky
	return math.Sqrt(dx*dx + dy*dy)
}

func (cr cheapruler) bearing(a []float64, b []float64) float64 {
	dx := (b[0] - a[0]) * cr.kx
	dy := (b[1] - a[1]) * cr.ky
	if dx == 0.0 && dy == 0.0 {
		return 0.0
	}
	var bearing = math.Atan2(dx, dy) * 180 / math.Pi
	if bearing > 180 {
		bearing -= 360
	}
	return bearing
}

func (cr cheapruler) destination(p []float64, dist float64, bearing float64) []float64 {
	a := (90.0 - bearing) * math.Pi / 180.0
	return cr.offset(p, math.Cos(a)*dist, math.Sin(a)*dist)
}

func (cr cheapruler) offset(p []float64, dx float64, dy float64) []float64 {
	xo := p[0] + dx/cr.kx
	yo := p[1] + dy/cr.ky
	return []float64{xo, yo}
}

func (cr cheapruler) lineDistance(points [][]float64) float64 {
	total := 0.0
	for i := 0; i < len(points)-1; i++ {
		total += cr.distance(points[i], points[i+1])
	}
	return total
}

func (cr cheapruler) area(polygon [][][]float64) float64 {
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

	return (math.Abs(sum) / 2) * cr.kx * cr.ky
}

func (cr cheapruler) along(line [][]float64, dist float64) []float64 {
	sum := 0.0

	if dist <= 0 {
		return line[0]
	}

	for i := 0; i < len(line)-1; i++ {
		var p0 = line[i]
		var p1 = line[i+1]
		var d = cr.distance(p0, p1)
		sum += d
		if sum > dist {
			return interpolate(p0, p1, (dist-(sum-d))/d)
		}
	}

	return line[len(line)-1]
}

func (cr cheapruler) pointOnLine(line [][]float64, p []float64) pol {
	minDist := math.Inf(1)
	var minX float64
	var minY float64
	var minI float64
	var minT float64
	var t float64

	for i := 0; i < len(line)-1; i++ {

		x := line[i][0]
		y := line[i][1]
		dx := (line[i+1][0] - x) * cr.kx
		dy := (line[i+1][1] - y) * cr.ky

		if dx != 0 || dy != 0 {

			t = ((p[0]-x)*cr.kx*dx + (p[1]-y)*cr.ky*dy) / (dx*dx + dy*dy)

			if t > 1 {
				x = line[i+1][0]
				y = line[i+1][1]

			} else if t > 0 {
				x += (dx / cr.kx) * t
				y += (dy / cr.ky) * t
			}
		}

		dx = (p[0] - x) * cr.kx
		dy = (p[1] - y) * cr.ky

		sqDist := dx*dx + dy*dy
		if sqDist < minDist {
			minDist = sqDist
			minX = x
			minY = y
			minI = float64(i)
			minT = t
		}
	}

	return pol{
		[]float64{minX, minY},
		minI,
		minT,
	}
}

func (cr cheapruler) lineSlice(start []float64, stop []float64, line [][]float64) [][]float64 {
	p1 := cr.pointOnLine(line, start)
	p2 := cr.pointOnLine(line, stop)

	if p1.index > p2.index || (p1.index == p2.index && p1.t > p2.t) {
		tmp := p1
		p1 = p2
		p2 = tmp
	}

	sl := [][]float64{p1.point}

	l := p1.index + 1
	r := p2.index

	if !equals(line[int(l)], sl[0]) && l <= r {
		sl = append(sl, line[int(l)])
	}

	for i := l + 1; i <= r; i++ {
		sl = append(sl, line[int(i)])
	}

	if !equals(line[int(r)], p2.point) {
		sl = append(sl, p2.point)
	}

	return sl
}

func (cr cheapruler) lineSliceAlong(start float64, stop float64, line [][]float64) [][]float64 {
	sum := 0.0
	var sl [][]float64

	for i := 0; i < len(line)-1; i++ {
		p0 := line[i]
		p1 := line[i+1]
		d := cr.distance(p0, p1)

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

func (cr cheapruler) bufferPoint(p []float64, buffer float64) []float64 {
	var v = buffer / cr.ky
	var h = buffer / cr.kx
	return []float64{
		p[0] - h,
		p[1] - v,
		p[0] + h,
		p[1] + v,
	}
}

func (cr cheapruler) bufferBBox(bbox []float64, buffer float64) []float64 {
	var v = buffer / cr.ky
	var h = buffer / cr.kx
	return []float64{
		bbox[0] - h,
		bbox[1] - v,
		bbox[2] + h,
		bbox[3] + v,
	}
}

func (cr cheapruler) insideBBox(p []float64, bbox []float64) bool {
	return p[0] >= bbox[0] &&
		p[0] <= bbox[2] &&
		p[1] >= bbox[1] &&
		p[1] <= bbox[3]
}

func equals(a []float64, b []float64) bool {
	return a[0] == b[0] && a[1] == b[1]
}

func interpolate(a []float64, b []float64, t float64) []float64 {
	var dx = b[0] - a[0]
	var dy = b[1] - a[1]
	return []float64{
		a[0] + dx*t,
		a[1] + dy*t,
	}
}
