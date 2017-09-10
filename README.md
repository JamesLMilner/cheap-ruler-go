# Cheapruler (Go)

[![godoc reference](godoc.png)](https://godoc.org/github.com/JamesMilnerUK/cheapruler-go)

"A collection of very fast approximations to common geodesic measurements. Useful for performance-sensitive code that measures things on a city scale.

The approximations are based on an [FCC-approved formula of ellipsoidal Earth projection](https://www.gpo.gov/fdsys/pkg/CFR-2005-title47-vol4/pdf/CFR-2005-title47-vol4-sec73-208.pdf).
For distances under 500 kilometers and not on the poles,
the results are very precise — within [0.1% margin of error](#precision)
compared to [Vincenti formulas](https://en.wikipedia.org/wiki/Vincenty%27s_formulae),
and usually much less for shorter distances."

## Usage

Here  is an example of doing a distance measurement in kilometers, with a ruler Latitude of 32.8351:

```go

cr, _ := NewCheapruler(32.8351, "kilometers")
pointA := []float64{-96.920341, 32.838261}
pointB := []float64{-96.920421, 32.838295}
dist := cr.Distance(pointA, pointB)
fmt.Println(dist)
// Output: 0.008385790760648736

```

## Acknowledgements

Based on [Vladimir Agafonkin](https://github.com/mourner)'s JavaScript library [cheapruler](https://github.com/mapbox/cheap-ruler)

## License 

ISC License