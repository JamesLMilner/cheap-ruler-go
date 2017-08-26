# Cheapruler (Go)

Based [Vladimir Agafonkin](https://github.com/mourner)'s JavaScript library [cheapruler](https://github.com/mapbox/cheap-ruler)

"A collection of very fast approximations to common geodesic measurements. Useful for performance-sensitive code that measures things on a city scale.

The approximations are based on an [FCC-approved formula of ellipsoidal Earth projection](https://www.gpo.gov/fdsys/pkg/CFR-2005-title47-vol4/pdf/CFR-2005-title47-vol4-sec73-208.pdf).
For distances under 500 kilometers and not on the poles,
the results are very precise — within [0.1% margin of error](#precision)
compared to [Vincenti formulas](https://en.wikipedia.org/wiki/Vincenty%27s_formulae),
and usually much less for shorter distances."

## License 

ISC License